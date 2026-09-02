package main

type World struct {
	SpawnRoom *Room
	Rooms     []*Room
	Players   map[string]*Player
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

func (w World) GetSpawnRoom() *Room {
	return w.SpawnRoom
}

func (w *World) AddRooms(rooms ...*Room) {
	w.Rooms = append(w.Rooms, rooms...)
}

func (w *World) AddPlayer(player *Player) {
	w.Players[player.Name] = player
}

func (w *World) SetSpawnRoom(room *Room) {
	w.SpawnRoom = room
}
