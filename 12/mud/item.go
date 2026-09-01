package main

type Item struct {
	Name    string
	Targets []*Target
}

func NewItem(name string) *Item {
	return &Item{
		Name: name,
	}
}

func (i Item) GetName() string {
	return i.Name
}

func (i Item) GetTargets() []*Target {
	return i.Targets
}

func (i *Item) AddTargets(elem ...*Target) {
	i.Targets = append(i.Targets, elem...)
}

func (i Item) CheckTarget(target string) (*Target, bool) {
	targets := i.GetTargets()
	for idx := range targets {
		if targets[idx].GetName() == target {
			return targets[idx], true
		}
	}
	return nil, false
}
