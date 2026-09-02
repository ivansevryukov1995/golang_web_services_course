package main

type Mission struct {
	Name     string
	Complete bool
}

func NewMission(name string) *Mission {
	return &Mission{
		Name:     name,
		Complete: false,
	}
}

func (m Mission) GetName() string {
	return m.Name
}

func (m *Mission) Completed() {
	m.Complete = true
}
