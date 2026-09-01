package main

import (
	"strings"
)

var (
	world *World
)

func main() {
	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все комнаты
		если что-то было - оно корректно перезатирается
	*/

	// Цели для взамодействия
	door := NewTarget("дверь")

	// Создание мелких предметов
	tea := NewItem("чай")
	keys := NewItem("ключи")
	notes := NewItem("конспекты")
	bag := NewItem("рюкзак")

	// Взаимодействия для мелких предметов
	keys.AddTargets(door)

	// Создание предметов окружения (envir)
	tableKitchen := NewEnvir("стол")
	tableRoom := NewEnvir("стол")
	chairRoom := NewEnvir("стул")

	// Наполнение предметов окружения (envir) мелкими предметами
	tableKitchen.AddItems(tea)
	tableRoom.AddItems(keys, notes)
	chairRoom.AddItems(bag)

	// Создание комнат
	kitchen := NewRoom("кухня", "кухня, ничего интересного")
	hall := NewRoom("коридор", "ничего интересного")
	room := NewRoom("комната", "ты в своей комнате")
	street := NewRoom("улица", "на улице весна")
	home := NewRoom("домой", "")

	// Поведение для комнат
	kitchen.SetKitchen()
	hall.AddTarget(door)
	street.AddTarget(door)

	// Наполнение комнат предметами окружения (envir)
	kitchen.AddEnvir(tableKitchen)
	room.AddEnvir(tableRoom, chairRoom)

	// Создание связей между комнатами
	kitchen.AddNextRooms(hall)
	hall.AddNextRooms(kitchen, room, street)
	room.AddNextRooms(hall)
	street.AddNextRooms(home)

	// Создание мира и наполнение его комнатами
	world = NewWorld()

	world.AddRooms(kitchen, hall, room, street, home)

	// Создание игрока
	playerOne := NewPlayer("Tristan", kitchen)

	// Цели игрока
	mission1 := NewMission("собрать рюкзак")
	mission2 := NewMission("идти в универ")

	playerOne.AddMissions(mission1, mission2)

	// Наполнение мира игроками
	world.AddPlayer(playerOne)
}

func handleCommand(command string) string {
	commands := strings.Split(command, " ")

	player := world.GetPlayer("Tristan")

	switch commands[0] {
	case "осмотреться":
		return player.Look()
	case "идти":
		return player.GoTo(commands[1])
	case "применить":
		return player.Apply(commands[1], commands[2])
	case "взять":
		return player.Take(commands[1])
	case "надеть":
		return player.PutOn(commands[1])
	default:
		return "неизвестная команда"
	}
}
