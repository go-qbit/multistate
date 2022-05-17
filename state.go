package multistate

type State interface {
	GetStateId() string
	Eval(v uint64) bool
}

type States []State

type state struct {
	id      string
	caption string
	bit     uint8
}

func (s *state) GetStateId() string {
	return s.id
}

func (s *state) Eval(v uint64) bool {
	return v&(1<<s.bit) > 0
}
