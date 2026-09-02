package main

import (
	"fmt"
	"strings"
)

type Player struct {
	Name        string
	HasBag      bool
	CurrentRoom *Room
	Missions    []*Mission
	ItemsPlayer []*Item
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:   name,
		HasBag: false,
	}
}

func (p Player) GetMissions() []*Mission {
	return p.Missions
}

func (p Player) GetItems() []*Item {
	return p.ItemsPlayer
}

func (p Player) GetCurrentRoom() *Room {
	return p.CurrentRoom
}

func (p *Player) UpdateCurrentRoom(room *Room) {
	p.CurrentRoom = room
}

func (p *Player) UpdateMission(missionName string) {
	missions := p.GetMissions()

	for idx := range missions {
		if missions[idx].GetName() == missionName {
			missions[idx].Completed()
			return
		}
	}
}

func (p *Player) AddSpawnRoom(room *Room) {
	p.CurrentRoom = room
}

func (p *Player) AddMissions(mission ...*Mission) {
	p.Missions = append(p.Missions, mission...)
}

func (p *Player) AddItems(items ...*Item) {
	p.ItemsPlayer = append(p.ItemsPlayer, items...)
}

func (p Player) CheckItem(itemName string) (*Item, bool) {
	items := p.GetItems()
	for idx := range items {
		if items[idx].GetName() == itemName {
			return items[idx], true
		}
	}
	return nil, false
}

func (p *Player) PutIn(item *Item) {
	p.ItemsPlayer = append(p.ItemsPlayer, item)
}

func (p *Player) Take(itemName string) string {
	if p.HasBag {
		item, ok := p.GetCurrentRoom().CheckItemInRoom(itemName)
		if !ok {
			return "нет такого"
		}
		p.PutIn(item)
		p.GetCurrentRoom().RemoveItemInRoom(item)

		return fmt.Sprintf("предмет добавлен в инвентарь: %s", itemName)
	} else {
		return "некуда класть"
	}
}

func (p *Player) PutOn(itemName string) string {
	switch itemName {
	case "рюкзак":
		bag, ok := p.GetCurrentRoom().CheckItemInRoom(itemName)
		if !ok {
			return ""
		}

		p.GetCurrentRoom().RemoveItemInRoom(bag)
		p.HasBag = true

		p.UpdateMission("собрать рюкзак")

		return fmt.Sprintf("вы надели: %s", bag.GetName())
	default:
		return "нет такого"
	}
}

func (p *Player) Apply(itemName, targetName string) string {
	item, ok := p.CheckItem(itemName)
	if !ok {
		return fmt.Sprintf("нет предмета в инвентаре - %s", itemName)
	}

	_, ok = p.GetCurrentRoom().CheckTarget(targetName)
	if !ok {
		return "не к чему применить"
	}

	target, ok := item.CheckTarget(targetName)
	if ok {
		target.Open()
		return targetName + " " + target.GetCondition()
	}

	return "нельзя применить"
}

// GetNextRoomsMsg вернет сообщение: куда можно идти из текущей комнаты
func (r Room) GetNextRoomsMsg() string {
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

func (p Player) Look() string {
	var msg []string

	// Текущая локация
	if p.GetCurrentRoom().IsKitchen() {
		msg = append(msg, "ты находишься на кухне")
	} else if p.GetCurrentRoom().IsEmptyRoom() {
		msg = append(msg, "пустая комната")
	}

	// Текущее окружение
	envirs := p.GetCurrentRoom().GetFurniture()
	var msgEnvirs []string

	for i := range envirs {
		items := envirs[i].GetItems()
		if len(items) == 0 {
			continue
		}

		var msgItems []string
		for j := range items {
			msgItems = append(msgItems, items[j].GetName())
		}

		if len(msgItems) != 0 {
			msgEnvirs = append(msgEnvirs, fmt.Sprintf("на %sе: %s", envirs[i].GetName(), strings.Join(msgItems, ", ")))
		}
	}

	if len(msgEnvirs) > 0 {
		msg = append(msg, strings.Join(msgEnvirs, ", "))
	}

	// Текущие миссии
	missions := p.GetMissions()
	if p.GetCurrentRoom().IsKitchen() && len(missions) != 0 {
		var activeMissions []string
		for _, m := range missions {
			if !m.Complete {
				activeMissions = append(activeMissions, m.GetName())
			}
		}
		if len(activeMissions) > 0 {
			msg = append(msg, fmt.Sprintf("надо %s", strings.Join(activeMissions, " и ")))
		}
	}

	onePart := strings.Join(msg, ", ")

	// Куда можно идти дальше
	twoPart := p.GetCurrentRoom().GetNextRoomsMsg()

	return strings.Join([]string{onePart, twoPart}, ". ")
}

func (p *Player) GoTo(roomName string) string {

	rooms := p.GetCurrentRoom().GetNextRooms()

	room := p.GetCurrentRoom().FindRoomByName(rooms, roomName)
	if room == nil {
		return fmt.Sprintf("нет пути в %s", roomName)
	}

	targets := p.GetCurrentRoom().GetTargets()
	for idx := range targets {
		_, ok := room.CheckTarget(targets[idx].GetName())
		if ok && targets[idx].GetCondition() == "закрыта" {
			return fmt.Sprintf("%s %s", targets[idx].GetName(), targets[idx].GetCondition())
		}
	}

	onePart := room.GetDesc()

	p.UpdateCurrentRoom(room)

	// Куда можно идти дальше
	twoPart := p.GetCurrentRoom().GetNextRoomsMsg()

	return strings.Join([]string{onePart, twoPart}, ". ")
}
