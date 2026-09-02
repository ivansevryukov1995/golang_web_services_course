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
	Messages    chan string
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:     name,
		HasBag:   false,
		Messages: make(chan string),
	}
}

func (p Player) GetMissions() []*Mission {
	return p.Missions
}

func (p Player) GetItems() []*Item {
	return p.ItemsPlayer
}

func (p *Player) GetCurrentRoom() *Room {
	return p.CurrentRoom
}

func (p *Player) GetName() string {
	return p.Name
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

func (p *Player) AddCurrentRoom(room *Room) {
	p.CurrentRoom = room
	room.AddNewPlayers(p)
}

func (p *Player) AddMissions(mission ...*Mission) {
	p.Missions = append(p.Missions, mission...)
}

func (p *Player) AddItems(items ...*Item) {
	p.ItemsPlayer = append(p.ItemsPlayer, items...)
}

func (p *Player) AddMessages(msg ...string) {
	for idx := range msg {
		p.Messages <- msg[idx]
	}

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

		return fmt.Sprintf("вы одели: %s", bag.GetName())
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

func (p *Player) Look() string {
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

	// Есть ли в комнате ещё кто-то кроме вас
	threePart := p.GetCurrentRoom().GetPlayersRoomMsg(p.GetName())

	if threePart != "" {
		twoPart += ". " + threePart
	}

	if twoPart != "" {
		return onePart + ". " + twoPart
	}

	return strings.Join([]string{onePart, twoPart}, ". ")
}

func (p *Player) GoTo(newRoomName string) string {
	oldRoom := p.GetCurrentRoom()

	nextRooms := oldRoom.GetNextRooms()

	newRoom := oldRoom.FindRoomByName(nextRooms, newRoomName)
	if newRoom == nil {
		return fmt.Sprintf("нет пути в %s", newRoomName)
	}

	targets := p.GetCurrentRoom().GetTargets()
	for idx := range targets {
		_, ok := newRoom.CheckTarget(targets[idx].GetName())
		if ok && targets[idx].GetCondition() == "закрыта" {
			return fmt.Sprintf("%s %s", targets[idx].GetName(), targets[idx].GetCondition())
		}
	}

	onePart := newRoom.GetDesc()

	p.UpdateCurrentRoom(newRoom)
	oldRoom.RemovePlayerInRoom(p)
	p.GetCurrentRoom().AddNewPlayers(p)

	// Куда можно идти дальше
	twoPart := p.GetCurrentRoom().GetNextRoomsMsg()

	return strings.Join([]string{onePart, twoPart}, ". ")
}

func (p *Player) Tell(msg string) {
	players := p.GetCurrentRoom().GetPlayers()
	for idx := range players {
		players[idx].AddMessages(p.GetName() + " говорит: " + msg)
	}
}

func (p *Player) TellSomeone(player string, msg string) {
	if msg == "" {
		msg = " выразительно молчит, смотря на вас"
	} else {
		msg = " говорит вам: " + msg
	}

	players := p.GetCurrentRoom().GetPlayers()

	for idx := range players {
		if players[idx].GetName() == player {
			players[idx].AddMessages(p.GetName() + msg)
			return
		}
	}

	p.AddMessages("тут нет такого игрока")
}

func (p *Player) HandleInput(command string) {
	commands := strings.Split(command, " ")
	var msg string

	switch commands[0] {
	case "осмотреться":
		p.AddMessages(p.Look())
	case "идти":
		p.AddMessages(p.GoTo(commands[1]))
	case "применить":
		p.AddMessages(p.Apply(commands[1], commands[2]))
	case "взять":
		p.AddMessages(p.Take(commands[1]))
	case "одеть":
		p.AddMessages(p.PutOn(commands[1]))
	case "сказать":
		msg = strings.Join(commands[1:], " ")
		p.Tell(msg)
	case "сказать_игроку":
		msg = strings.Join(commands[2:], " ")
		p.TellSomeone(commands[1], msg)
	default:
		p.AddMessages("неизвестная команда")
	}
}

func (p *Player) GetOutput() chan string {
	return p.Messages
}
