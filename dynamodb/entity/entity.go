package entity

type Entity struct {
	data interface{}
}

func New(data interface{}) *Entity {
	return &Entity{data}
}

func (e *Entity) MarshalJSON() ([]byte, error) {
	return Marshal(e.data)
}
