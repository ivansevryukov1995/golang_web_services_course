package main

type Room struct {
	Name               string
	Desc               string
	isKitchen          bool
	TargetsInteraction []*Target
	Envir              []*Envir
	NextRooms          []*Room
}

func NewRoom(name, desc string) *Room {
	return &Room{
		Name:               name,
		Desc:               desc,
		isKitchen:          false,
		TargetsInteraction: []*Target{},
		NextRooms:          []*Room{},
	}
}

func (r Room) GetEnvir() []*Envir {
	return r.Envir
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

func (r Room) GetTargets() []*Target {
	return r.TargetsInteraction
}

func (r *Room) SetKitchen() {
	r.isKitchen = true
}
func (r Room) IsKitchen() bool {
	return r.isKitchen
}

func (r Room) IsEmptyRoom() bool {
	envirs := r.GetEnvir()
	items := []*Item{}

	for idx := range envirs {
		items = append(items, envirs[idx].GetItems()...)
	}

	if len(items) != 0 {
		return true
	}
	return false
}

func (r *Room) AddEnvir(envirs ...*Envir) {
	r.Envir = append(r.Envir, envirs...)
}

func (r *Room) AddTarget(target *Target) {
	r.TargetsInteraction = append(r.TargetsInteraction, target)
}

func (r *Room) AddNextRooms(rooms ...*Room) {
	r.NextRooms = append(r.NextRooms, rooms...)
}

func (r *Room) CheckTarget(target string) (*Target, bool) {
	targets := r.GetTargets()
	for idx := range targets {
		if targets[idx].GetName() == target {
			return targets[idx], true
		}
	}
	return nil, false
}

func (r Room) CheckItemInRoom(item string) (*Item, bool) {
	envirs := r.GetEnvir()
	for idx := range envirs {
		item, ok := envirs[idx].CheckItemInEnvir(item)
		if ok {
			return item, true
		}
	}
	return nil, false
}

func (r *Room) RemoveItemInRoom(item *Item) {
	envirs := r.GetEnvir()
	for idx := range envirs {
		envirs[idx].RemoveItemInEnvir(item)
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
