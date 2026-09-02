package main

var (
	world *World
)

func main() {
	// initGame()

	// scanner := bufio.NewScanner(os.Stdin)
	// fmt.Println("Игра началась. Введите команду (или 'выход' для завершения):")

	// for {
	// 	fmt.Print("> ")
	// 	if !scanner.Scan() {
	// 		break
	// 	}

	// 	command := strings.TrimSpace(scanner.Text())
	// 	if command == "" {
	// 		continue
	// 	}
	// 	if command == "выход" || command == "quit" {
	// 		fmt.Println("Игра окончена.")
	// 		break
	// 	}

	// 	fmt.Println(handleCommand(command))
	// }

	// if err := scanner.Err(); err != nil {
	// 	fmt.Fprintln(os.Stderr, "ошибка ввода:", err)
	// }
}

func addPlayer(player *Player) {
	// Цели игрока
	mission1 := NewMission("собрать рюкзак")
	mission2 := NewMission("идти в универ")
	player.AddMissions(mission1, mission2)

	player.AddCurrentRoom(world.GetSpawnRoom())

	world.AddPlayer(player)
}

func initGame() {
	// Создание мелких предметов
	tea := NewItem("чай")
	keys := NewItem("ключи")
	notes := NewItem("конспекты")
	bag := NewItem("рюкзак")

	// Создание предметов окружения (Furniture)
	tableKitchen := NewFurniture("стол")
	tableRoom := NewFurniture("стол")
	chairRoom := NewFurniture("стул")

	// Наполнение предметов окружения (Furniture) мелкими предметами
	tableKitchen.AddItems(tea)
	tableRoom.AddItems(keys, notes)
	chairRoom.AddItems(bag)

	// Создание комнат
	kitchen := NewRoom("кухня", "кухня, ничего интересного")
	hall := NewRoom("коридор", "ничего интересного")
	room := NewRoom("комната", "ты в своей комнате")
	street := NewRoom("улица", "на улице весна")
	home := NewRoom("домой", "")

	// Создание дверей
	door := NewDoor("дверь")

	// Взаимодействия для мелких предметов
	keys.AddTargets(door)

	// Поведение для комнат
	kitchen.SetKitchen()
	hall.AddTarget(door)
	street.AddTarget(door)

	// Наполнение комнат предметами окружения (Furniture)
	kitchen.AddFurniture(tableKitchen)
	room.AddFurniture(tableRoom, chairRoom)

	// Создание связей между комнатами
	kitchen.AddNextRooms(hall)
	hall.AddNextRooms(kitchen, room, street)
	room.AddNextRooms(hall)
	street.AddNextRooms(home)

	// Создание мира и наполнение его комнатами
	world = NewWorld()

	world.AddRooms(kitchen, hall, room, street, home)

	world.SetSpawnRoom(kitchen)

}

// func handleCommand(command string) string {
// 	commands := strings.Split(command, " ")

// 	player := world.GetPlayer("Tristan")

// 	switch commands[0] {
// 	case "осмотреться":
// 		return player.Look()
// 	case "идти":
// 		return player.GoTo(commands[1])
// 	case "применить":
// 		return player.Apply(commands[1], commands[2])
// 	case "взять":
// 		return player.Take(commands[1])
// 	case "надеть":
// 		return player.PutOn(commands[1])
// 	default:
// 		return "неизвестная команда"
// 	}
// }
