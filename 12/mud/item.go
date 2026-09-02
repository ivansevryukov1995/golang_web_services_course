package main

type Item struct {
	Name    string
	Targets []*Door
}

func NewItem(name string) *Item {
	return &Item{
		Name: name,
	}
}

func (i Item) GetName() string {
	return i.Name
}

func (i Item) GetTargets() []*Door {
	return i.Targets
}

func (i *Item) AddTargets(elem ...*Door) {
	i.Targets = append(i.Targets, elem...)
}

func (i Item) CheckTarget(target string) (*Door, bool) {
	targets := i.GetTargets()
	for idx := range targets {
		if targets[idx].GetName() == target {
			return targets[idx], true
		}
	}
	return nil, false
}
