package main

type Door struct {
	Name        string
	Interaction string
}

func NewDoor(name string) *Door {
	return &Door{
		Name:        name,
		Interaction: "закрыта",
	}
}

func (t Door) GetName() string {
	return t.Name
}

func (t Door) GetCondition() string {
	return t.Interaction
}

func (t *Door) Close() {
	t.Interaction = "закрыта"
}

func (t *Door) Open() {
	t.Interaction = "открыта"
}
