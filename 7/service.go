package main

import (
	"context"
	"encoding/json"
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

type StatEvent struct {
	Method   string
	Consumer string
}

// statSnapshot — локальные счётчики одного подписчика
type statSnapshot struct {
	mu         sync.Mutex
	byMethod   map[string]uint64
	byConsumer map[string]uint64
}

func newStatSnapshot() *statSnapshot {
	return &statSnapshot{
		byMethod:   make(map[string]uint64),
		byConsumer: make(map[string]uint64),
	}
}

func (ss *statSnapshot) handle(e StatEvent) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.byMethod[e.Method]++
	ss.byConsumer[e.Consumer]++
}

// flushAndReset возвращает накопленные данные и обнуляет счётчики
func (ss *statSnapshot) flushAndReset() (method, consumer map[string]uint64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	method = ss.byMethod
	consumer = ss.byConsumer

	ss.byMethod = make(map[string]uint64)
	ss.byConsumer = make(map[string]uint64)

	return
}

type EventBus struct {
	mu   sync.RWMutex
	subs []*statSnapshot
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe создаёт новый snapshot и регистрирует его в шине
func (eb *EventBus) Subscribe() *statSnapshot {
	snap := newStatSnapshot()

	eb.mu.Lock()
	eb.subs = append(eb.subs, snap)
	eb.mu.Unlock()

	return snap
}

// Unsubscribe удаляет snapshot из шины
func (eb *EventBus) Unsubscribe(snap *statSnapshot) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for i, s := range eb.subs {
		if s == snap {
			eb.subs = append(eb.subs[:i], eb.subs[i+1:]...)
			return
		}
	}
}

// Publish рассылает событие всем подписчикам
func (eb *EventBus) Publish(e StatEvent) {
	eb.mu.RLock()
	subs := make([]*statSnapshot, len(eb.subs))
	copy(subs, eb.subs)
	eb.mu.RUnlock()

	for _, snap := range subs {
		snap.handle(e)
	}
}

type configACL map[string][]string

type server struct {
	UnimplementedAdminServer
	UnimplementedBizServer

	mu           sync.Mutex
	aclConfig    configACL
	observerList []chan *Event

	statBus *EventBus
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

func HasAccess(allowed []string, method string) bool {
	newPrefix := strings.Builder{}

	for _, pattern := range allowed {
		if pattern == "*" {
			return true
		}

		prefix, ok := strings.CutSuffix(pattern, "/*")
		if ok {
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

func CheckACL(
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

	if HasAccess(allowed, method) {
		return nil
	}

	return status.Error(codes.Unauthenticated, "nothing access")
}

func (s *server) ACLUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (
		resp any,
		err error,
	) {

		if err = CheckACL(ctx, info.FullMethod, s.aclConfig); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (s *server) ACLStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		if err := CheckACL(ss.Context(), info.FullMethod, s.aclConfig); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

func NewEvent(ctx context.Context, method string) *Event {
	md, _ := metadata.FromIncomingContext(ctx)

	// Получение consumer
	roles := md.Get("consumer")
	consumer := roles[0]

	// Получение host
	var host string
	if p, ok := peer.FromContext(ctx); ok {
		host = p.Addr.String()
	}

	return &Event{
		Timestamp: time.Now().Unix(),
		Consumer:  consumer,
		Method:    method,
		Host:      host,
	}
}

func (s *server) NotifyAll(event *Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range s.observerList {
		ch <- event
	}
}

func (s *server) EventSendUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (
		resp any,
		err error) {

		s.NotifyAll(NewEvent(ctx, info.FullMethod))

		return handler(ctx, req)
	}
}

func (s *server) EventSendStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		s.NotifyAll(NewEvent(ss.Context(), info.FullMethod))

		return handler(srv, ss)
	}
}

func ExtractConsumer(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)

	// Получение consumer
	roles := md.Get("consumer")
	consumer := roles[0]

	return consumer
}

func (s *server) StatSendUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		s.statBus.Publish(StatEvent{
			Method:   info.FullMethod,
			Consumer: ExtractConsumer(ctx),
		})

		return handler(ctx, req)
	}
}

func (s *server) StatSendStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		s.statBus.Publish(StatEvent{
			Method:   info.FullMethod,
			Consumer: ExtractConsumer(ss.Context()),
		})

		return handler(srv, ss)
	}
}

func (s *server) Register() chan *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *Event)
	s.observerList = append(s.observerList, ch)

	return ch
}

func (s *server) Deregister(target chan *Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, ch := range s.observerList {
		if ch == target {
			s.observerList = append(s.observerList[:i], s.observerList[i+1:]...)
			close(target)
			break
		}
	}
}

func (s *server) Logging(_ *Nothing, stream Admin_LoggingServer) error {
	ch := s.Register()
	defer s.Deregister(ch)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event := <-ch:
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

func (s *server) Statistics(req *StatInterval, stream Admin_StatisticsServer) error {
	// Каждый клиент получает свою локальную подписку
	snap := s.statBus.Subscribe()
	defer s.statBus.Unsubscribe(snap)

	ticker := time.NewTicker(time.Duration(req.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			// Берём накопленные данные и сразу обнуляем — это и есть дифф
			byMethod, byConsumer := snap.flushAndReset()

			if err := stream.Send(&Stat{
				ByMethod:   byMethod,
				ByConsumer: byConsumer,
			}); err != nil {
				return err
			}
		}
	}
}

func ParseACL(ACLData string) (configACL, error) {
	var conf configACL

	if err := json.Unmarshal([]byte(ACLData), &conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) error {
	configACL, err := ParseACL(ACLData)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	srv := &server{
		aclConfig: configACL,
		statBus:   NewEventBus(),
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srv.ACLUnaryInterceptor(),
			srv.EventSendUnaryInterceptor(),
			srv.StatSendUnaryInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			srv.ACLStreamInterceptor(),
			srv.EventSendStreamInterceptor(),
			srv.StatSendStreamInterceptor(),
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
