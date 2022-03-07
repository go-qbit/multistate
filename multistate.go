package multistate

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/go-qbit/qerror"
	"github.com/go-qbit/rbac"

	"github.com/go-qbit/multistate/expr"
)

var reStateAction = regexp.MustCompile(`^[a-z0-9_-]+$`)

type Multistate struct {
	emptyStateName string
	statesMap      map[string]state
	statesBitsMap  map[uint8]state
	actionsMap     map[string]action
	statesActions  map[uint64]map[string]uint64
}

type StateFlag struct {
	Id      string
	Bit     uint8
	Caption string
}

func New(emptyStateName string) *Multistate {
	return &Multistate{
		emptyStateName: emptyStateName,
		statesMap:      make(map[string]state),
		statesBitsMap:  make(map[uint8]state),
		actionsMap:     make(map[string]action),
	}
}

func (m *Multistate) AddState(bit uint8, id, caption string) state {
	if !reStateAction.MatchString(id) {
		panic(qerror.Errorf("Invalid characters in state id '%s', must be %s", id, reStateAction.String()))
	}

	if bit > 63 {
		panic(qerror.Errorf("Bit must be less than 64", id))
	}

	if _, exists := m.statesMap[id]; id == "empty" || id == "any" || exists {
		panic(qerror.Errorf("State '%s' already exists", id))
	}

	if _, exists := m.statesBitsMap[bit]; exists {
		panic(qerror.Errorf("Bit '%d' already busy", bit))
	}

	s := state{id, caption, bit}
	m.statesMap[id] = s

	return s
}

func (m *Multistate) AddAction(id, caption string, from expr.IExpression, set, reset []IState, onDo ActionDoFunc, permission *rbac.Permission) {
	if !reStateAction.MatchString(id) {
		panic(qerror.Errorf("Invalid characters in action id '%s', must be %s", id, reStateAction.String()))
	}

	a := action{
		id:         id,
		caption:    caption,
		from:       from,
		set:        make([]uint64, len(set)),
		reset:      make([]uint64, len(reset)),
		do:         onDo,
		permission: permission,
	}

	for i, s := range set {
		if state, exists := m.statesMap[s.GetStateId()]; exists {
			a.set[i] = 1 << state.bit
		} else {
			panic(qerror.Errorf("State '%s' doesn't exists", s.GetStateId()))
		}
	}

	for i, s := range reset {
		if state, exists := m.statesMap[s.GetStateId()]; exists {
			a.reset[i] = ^(1 << state.bit)
		} else {
			panic(qerror.Errorf("State '%s' doesn't exists", s.GetStateId()))
		}
	}

	m.actionsMap[id] = a
}

func (m *Multistate) Compile() {
	if m.statesActions != nil {
		panic(qerror.Errorf("Multistate is already compiled"))
	}

	m.statesActions = make(map[uint64]map[string]uint64)
	m.statesActions[0] = make(map[string]uint64)

	changed := true
	for changed {
		//println()
		changed = false

		for _, action := range m.actionsMap {
			for state, actions := range m.statesActions {
				if action.from.Eval(state) {
					if _, exists := actions[action.id]; !exists {
						changed = true
						newState := state
						for _, v := range action.set {
							//println("set", v)
							newState |= v
						}
						for _, v := range action.reset {
							newState &= v
							//println("reset", v)
						}

						//println("from", state, "to", newState, "by", action.id)

						actions[action.id] = newState
						if _, exists := m.statesActions[newState]; !exists {
							m.statesActions[newState] = make(map[string]uint64)
						}
					}
				}
			}
		}
	}
}

func (m *Multistate) GetStateActions(ctx context.Context, state uint64) []string {
	if actions, exists := m.statesActions[state]; exists {
		res := make([]string, 0, len(actions))

		for actionId := range actions {
			action := m.actionsMap[actionId]
			if action.permission == nil || rbac.HasPermission(ctx, action.permission) {
				res = append(res, actionId)
			}
		}

		return res
	}

	return nil
}

func (m *Multistate) DoAction(ctx context.Context, model IModel, action string, opts ...interface{}) (uint64, error) {
	ctx, err := model.StartAction(ctx)
	if err != nil {
		return 0, model.EndAction(ctx, err)
	}

	curState, err := model.GetState(ctx)
	if err != nil {
		return 0, model.EndAction(ctx, err)
	}

	actions, exists := m.statesActions[curState]
	if !exists {
		return 0, model.EndAction(ctx, qerror.Errorf("invalid state %d", curState))
	}

	newState, exists := actions[action]
	if !exists {
		return 0, model.EndAction(ctx, qerror.Errorf("invalid action '%s' state %d", action, curState))
	}

	if onAction := m.actionsMap[action].do; onAction != nil {
		if err := onAction(ctx, model, opts...); err != nil {
			return 0, model.EndAction(ctx, err)
		}
	}

	if err := model.SetState(ctx, newState); err != nil {
		return 0, model.EndAction(ctx, err)
	}

	return newState, model.EndAction(ctx, nil)
}

func (m *Multistate) GetStateFlags(id uint64) []StateFlag {
	if id == 0 {
		return []StateFlag{}
	}

	var res []StateFlag
	for _, state := range m.statesMap {
		if id&(1<<state.bit) > 0 {
			res = append(res, StateFlag{
				Id:      state.id,
				Bit:     state.bit,
				Caption: state.caption,
			})
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Bit < res[j].Bit
	})

	return res
}

func (m *Multistate) GetStateName(id uint64) string {
	flags := m.GetStateFlags(id)

	if len(flags) == 0 {
		if m.emptyStateName == "" {
			return "empty"
		}
		return m.emptyStateName
	}

	stateNames := make([]string, len(flags))
	for i, flag := range flags {
		stateNames[i] = flag.Caption
	}

	return strings.Join(stateNames, ".\n") + "."
}

func (m *Multistate) GetActionName(id string) string {
	return m.actionsMap[id].caption
}
