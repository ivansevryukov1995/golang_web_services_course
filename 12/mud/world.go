package main

type World struct {
	Rooms   []*Room
	Players map[string]*Player
}

func NewWorld() *World {
	return &World{
		Rooms:   []*Room{},
		Players: make(map[string]*Player),
	}
}

func (w World) GetPlayer(name string) *Player {
	return w.Players[name]
}

func (w *World) AddRooms(rooms ...*Room) {
	w.Rooms = append(w.Rooms, rooms...)
}

func (w *World) AddPlayer(p *Player) {
	w.Players[p.Name] = p
}
