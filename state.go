package multistate

type IState interface {
	GetStateId() string
}

type state struct {
	id      string
	caption string
	bit     uint8
}

func (s state) GetStateId() string {
	return s.id
}

func (s state) Eval(v uint64) bool {
	return v&(1<<s.bit) > 0
}