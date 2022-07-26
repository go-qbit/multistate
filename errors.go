package multistate

import "errors"

var (
	ErrInvalidState    = errors.New("invalid_state_error")
	ErrInvalidAction   = errors.New("invalid_action_error")
	ErrExecutionAction = errors.New("execute_action_error")
	ErrSetState        = errors.New("set_state_error")
)
