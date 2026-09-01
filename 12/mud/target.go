package main

type Target struct {
	Name        string
	Interaction string
}

func NewTarget(name string) *Target {
	return &Target{
		Name:        name,
		Interaction: "закрыта",
	}
}

func (t Target) GetName() string {
	return t.Name
}

func (t Target) GetInteraction() string {
	return t.Interaction
}

func (t *Target) UpdateInteraction(s string) {
	t.Interaction = s
}
