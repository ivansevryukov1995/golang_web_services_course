package main

type Furniture struct {
	Name  string
	Items []*Item
}

func NewFurniture(name string) *Furniture {
	return &Furniture{
		Name:  name,
		Items: []*Item{},
	}
}

func (e Furniture) GetName() string {
	return e.Name
}

func (e Furniture) GetItems() []*Item {
	return e.Items
}

func (e *Furniture) AddItems(items ...*Item) {
	e.Items = append(e.Items, items...)
}

func (e Furniture) CheckItemInFurniture(itemName string) (*Item, bool) {
	items := e.GetItems()
	for idx := range items {
		if items[idx].GetName() == itemName {
			return items[idx], true
		}
	}
	return nil, false
}

func (e *Furniture) RemoveItemInFurniture(item *Item) {
	items := e.GetItems()
	for idx := range items {
		if items[idx] == item {
			items[idx] = nil
			e.Items = append(e.Items[:idx], e.Items[idx+1:]...)
			return
		}
	}
}
