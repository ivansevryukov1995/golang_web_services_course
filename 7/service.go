package main

import (
	"context"
	"encoding/json"
	"maps"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type configACL map[string][]string

type server struct {
	UnimplementedAdminServer
	UnimplementedBizServer

	mu             sync.Mutex
	aclConfig      configACL
	observersEvent []chan Event

	byMethod   map[string]uint64
	byConsumer map[string]uint64
}

func (s *server) Logging(_ *Nothing, stream Admin_LoggingServer) error {
	ch := s.attachEvent()
	defer s.detachEvent(ch)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event := <-ch:
			if err := stream.Send(&event); err != nil {
				return err
			}
		}
	}
}

func (s *server) Statistics(req *StatInterval, stream Admin_StatisticsServer) error {
	prevByMethod, prevByConsumer := s.snapshotStat()

	ticker := time.NewTicker(time.Duration(req.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			diff := s.diffStat(prevByMethod, prevByConsumer)
			if err := stream.Send(diff); err != nil {
				return err
			}
			// Обновляем базу для следующего интервала
			prevByMethod, prevByConsumer = s.snapshotStat()
		}
	}
}

func (s *server) Check(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *server) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *server) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *server) aclUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (
		resp any,
		err error,
	) {

		if err = checkACL(ctx, info.FullMethod, s.aclConfig); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (s *server) aclStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		if err := checkACL(ss.Context(), info.FullMethod, s.aclConfig); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

func (s *server) eventSendUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (
		resp any,
		err error) {

		s.notifyEvent(newEvent(ctx, info.FullMethod))

		return handler(ctx, req)
	}
}

func (s *server) eventSendStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		s.notifyEvent(newEvent(ss.Context(), info.FullMethod))

		return handler(srv, ss)
	}
}
func (s *server) statSendUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (
		resp any,
		err error) {

		s.incStat(info.FullMethod, extractConsumer(ctx))

		return handler(ctx, req)
	}
}

func (s *server) statSendStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		s.incStat(info.FullMethod, extractConsumer(ss.Context()))

		return handler(srv, ss)
	}
}

func (s *server) notifyEvent(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range s.observersEvent {
		select {
		case ch <- event: // неблокирующая отправка
		default:
		}
	}
}

func (s *server) attachEvent() chan Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan Event)
	s.observersEvent = append(s.observersEvent, ch)

	return ch
}

func (s *server) detachEvent(target chan Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, ch := range s.observersEvent {
		if ch == target {
			s.observersEvent = append(s.observersEvent[:i], s.observersEvent[i+1:]...)
			close(target)
			break
		}
	}
}

func (s *server) incStat(method, consumer string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.byMethod == nil {
		s.byMethod = make(map[string]uint64)
		s.byConsumer = make(map[string]uint64)
	}

	s.byMethod[method]++
	s.byConsumer[consumer]++
}

func (s *server) snapshotStat() (map[string]uint64, map[string]uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	byMethod := make(map[string]uint64, len(s.byMethod))
	maps.Copy(byMethod, s.byMethod)

	byConsumer := make(map[string]uint64, len(s.byConsumer))
	maps.Copy(byConsumer, s.byConsumer)

	return byMethod, byConsumer
}

func (s *server) diffStat(prevByMethod, prevByConsumer map[string]uint64) *Stat {
	curByMethod, curByConsumer := s.snapshotStat()

	res := &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}

	for k, v := range curByMethod {
		if v > prevByMethod[k] {
			res.ByMethod[k] = v - prevByMethod[k]
		}
	}

	for k, v := range curByConsumer {
		if v > prevByConsumer[k] {
			res.ByConsumer[k] = v - prevByConsumer[k]
		}
	}

	return res
}

func parseACL(ACLData string) (configACL, error) {
	var conf configACL

	if err := json.Unmarshal([]byte(ACLData), &conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func hasAccess(allowed []string, method string) bool {
	newPrefix := strings.Builder{}

	for _, pattern := range allowed {
		if pattern == "*" {
			return true
		}

		if strings.HasSuffix(pattern, "/*") {
			// Достаем префикс
			prefix := strings.TrimSuffix(pattern, "/*")

			newPrefix.Reset()
			newPrefix.WriteString(prefix)
			newPrefix.WriteString("/")

			if strings.HasPrefix(method, newPrefix.String()) {
				return true
			}
			continue
		}

		if pattern == method {
			return true
		}
	}
	return false
}

func checkACL(
	ctx context.Context,
	method string,
	conf configACL) (err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Получение consumer
	roles := md.Get("consumer")
	if len(roles) == 0 {
		return status.Error(codes.Unauthenticated, "missing consumer")
	}
	consumer := roles[0]

	allowed, exist := conf[consumer]
	if !exist {
		return status.Error(codes.Unauthenticated, "unknown consumer")
	}

	if hasAccess(allowed, method) {
		return nil
	}

	return status.Error(codes.Unauthenticated, "nothing access")
}

func newEvent(ctx context.Context, method string) Event {
	md, _ := metadata.FromIncomingContext(ctx)

	// Получение consumer
	roles := md.Get("consumer")
	consumer := roles[0]

	// Получение host
	var host string
	if p, ok := peer.FromContext(ctx); ok {
		host = p.Addr.String()
	}

	return Event{
		Timestamp: time.Now().Unix(),
		Consumer:  consumer,
		Method:    method,
		Host:      host,
	}
}

func extractConsumer(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)

	// Получение consumer
	roles := md.Get("consumer")
	consumer := roles[0]

	return consumer
}

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) error {
	ACLConfig, err := parseACL(ACLData)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	srv := &server{
		aclConfig: ACLConfig,
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srv.aclUnaryInterceptor(),
			srv.eventSendUnaryInterceptor(),
			srv.statSendUnaryInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			srv.aclStreamInterceptor(),
			srv.eventSendStreamInterceptor(),
			srv.statSendStreamInterceptor(),
		),
	)

	RegisterAdminServer(grpcServer, srv)
	RegisterBizServer(grpcServer, srv)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			return
		}
	}()

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	return err
}
