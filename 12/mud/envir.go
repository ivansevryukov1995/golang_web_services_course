package main

type Envir struct {
	Name  string
	Items []*Item
}

func NewEnvir(name string) *Envir {
	return &Envir{
		Name:  name,
		Items: []*Item{},
	}
}

func (e Envir) GetName() string {
	return e.Name
}
func (e Envir) GetItems() []*Item {
	return e.Items
}

func (e *Envir) AddItems(items ...*Item) {
	e.Items = append(e.Items, items...)
}

func (e Envir) CheckItemInEnvir(itemName string) (*Item, bool) {
	items := e.GetItems()
	for idx := range items {
		if items[idx].GetName() == itemName {
			return items[idx], true
		}
	}
	return nil, false
}

func (e *Envir) RemoveItemInEnvir(item *Item) {
	for idx := range e.GetItems() {
		if e.Items[idx].GetName() == item.GetName() {
			e.Items = append(e.Items[:idx], e.Items[idx+1:]...)
			return
		}
	}
}
