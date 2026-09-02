package main

import (
	"fmt"
	"strings"
)

type Room struct {
	Name      string
	Desc      string
	isKitchen bool
	Doors     []*Door
	Furniture []*Furniture
	NextRooms []*Room
	Players   []*Player
}

func NewRoom(name, desc string) *Room {
	return &Room{
		Name:      name,
		Desc:      desc,
		isKitchen: false,
		Doors:     []*Door{},
		NextRooms: []*Room{},
	}
}

func (r Room) GetFurniture() []*Furniture {
	return r.Furniture
}

func (r Room) GetName() string {
	return r.Name
}

func (r Room) GetNextRooms() []*Room {
	return r.NextRooms
}

func (r Room) GetDesc() string {
	return r.Desc
}

func (r Room) GetTargets() []*Door {
	return r.Doors
}

func (r Room) GetPlayers() []*Player {
	return r.Players
}

func (r *Room) SetKitchen() {
	r.isKitchen = true
}
func (r Room) IsKitchen() bool {
	return r.isKitchen
}

func (r Room) IsEmptyRoom() bool {
	envirs := r.GetFurniture()

	for idx := range envirs {
		if len(envirs[idx].GetItems()) > 0 {
			return false
		}
	}

	return true
}

func (r *Room) AddFurniture(envirs ...*Furniture) {
	r.Furniture = append(r.Furniture, envirs...)
}

func (r *Room) AddTarget(target *Door) {
	r.Doors = append(r.Doors, target)
}

func (r *Room) AddNextRooms(rooms ...*Room) {
	r.NextRooms = append(r.NextRooms, rooms...)
}

func (r *Room) AddNewPlayers(players ...*Player) {
	r.Players = append(r.Players, players...)
}

func (r *Room) CheckTarget(target string) (*Door, bool) {
	targets := r.GetTargets()
	for idx := range targets {
		if targets[idx].GetName() == target {
			return targets[idx], true
		}
	}
	return nil, false
}

func (r Room) CheckItemInRoom(item string) (*Item, bool) {
	envirs := r.GetFurniture()
	for idx := range envirs {
		item, ok := envirs[idx].CheckItemInFurniture(item)
		if ok {
			return item, true
		}
	}
	return nil, false
}

func (r *Room) RemoveItemInRoom(item *Item) {
	envirs := r.GetFurniture()
	for idx := range envirs {
		envirs[idx].RemoveItemInFurniture(item)
	}
}

func (r *Room) RemovePlayerInRoom(player *Player) {
	players := r.GetPlayers()
	for idx := range players {
		if players[idx] == player {
			players[idx] = nil
			r.Players = append(players[:idx], players[idx+1:]...)
			return
		}
	}
}

func (r Room) FindRoomByName(rooms []*Room, elemName string) *Room {
	for idx := range rooms {
		if rooms[idx].GetName() == elemName {
			return rooms[idx]
		}
	}
	return nil
}

// GetNextRoomsMsg вернет сообщение: куда можно идти из текущей комнаты
func (r *Room) GetNextRoomsMsg() string {
	rooms := r.GetNextRooms()
	if len(rooms) == 0 {
		return ""
	}
	names := make([]string, 0, len(rooms))
	for idx := range rooms {
		names = append(names, rooms[idx].GetName())
	}
	return fmt.Sprintf("можно пройти - %s", strings.Join(names, ", "))
}

// GetPlayersRoomMsg вернет сообщение: кто ещё из игроков находится в текущей комнате, кроме вас(asking)
func (r *Room) GetPlayersRoomMsg(asking string) string {
	players := r.GetPlayers()
	if len(players) == 1 {
		return ""
	}
	names := make([]string, 0, len(players))
	for idx := range players {
		if players[idx].GetName() != asking {
			names = append(names, players[idx].GetName())
		}
	}
	return fmt.Sprintf("Кроме вас тут ещё %s", strings.Join(names, ", "))
}
