package main

type Room struct {
	Name      string
	Desc      string
	isKitchen bool
	Doors     []*Door
	Furniture []*Furniture
	NextRooms []*Room
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

func (r Room) FindRoomByName(rooms []*Room, elemName string) *Room {
	for idx := range rooms {
		if rooms[idx].GetName() == elemName {
			return rooms[idx]
		}
	}
	return nil
}
