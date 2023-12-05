package multistate

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/go-qbit/multistate/expr"
)

var reStateAction = regexp.MustCompile(`^[a-z\d_-]+$`)

type Multistate struct {
	emptyStateName  string
	statesMap       map[string]*state
	statesBitsMap   map[uint8]*state
	actionsMap      map[string]*action
	statesActions   map[uint64]map[string]uint64
	clusters        []cluster
	stateClusterMap map[uint64]*cluster
	onDo            OnDoCallback
}

type StateFlag struct {
	Id      string
	Bit     uint8
	Caption string
}

type OnDoCallback func(ctx context.Context, entity Entity, prevState, newState uint64, action string, opts ...interface{}) error

type cluster struct {
	id         uint8
	name       string
	expression expr.Expression
}

func New(emptyStateName string) *Multistate {
	return &Multistate{
		emptyStateName: emptyStateName,
		statesMap:      make(map[string]*state),
		statesBitsMap:  make(map[uint8]*state),
		actionsMap:     make(map[string]*action),
	}
}

func (m *Multistate) SetOnDoCallback(cb OnDoCallback) {
	m.onDo = cb
}

func (m *Multistate) AddState(bit uint8, id, caption string) (*state, error) {
	if !reStateAction.MatchString(id) {
		return nil, fmt.Errorf("invalid characters in state id '%s', must be %s", id, reStateAction.String())
	}

	if bit > 63 {
		return nil, fmt.Errorf("bit must be less than 64")
	}

	if _, exists := m.statesMap[id]; id == "empty" || id == "any" || exists {
		return nil, fmt.Errorf("state '%s' already exists", id)
	}

	if _, exists := m.statesBitsMap[bit]; exists {
		return nil, fmt.Errorf("bit '%d' already busy", bit)
	}

	s := &state{id, caption, bit}
	m.statesMap[id] = s
	m.statesBitsMap[bit] = s

	return s, nil
}

func (m *Multistate) MustAddState(bit uint8, id, caption string) *state {
	s, err := m.AddState(bit, id, caption)
	if err != nil {
		panic(err)
	}

	return s
}

func (m *Multistate) AddAction(id, caption string, from expr.Expression, set, reset States, onDo ActionDoFunc, avail Availabler) error {
	if !reStateAction.MatchString(id) {
		return fmt.Errorf("invalid characters in action id '%s', must be %s", id, reStateAction.String())
	}

	if _, exists := m.actionsMap[id]; exists {
		return fmt.Errorf("action '%s' already exists", id)
	}

	a := &action{
		id:         id,
		caption:    caption,
		from:       from,
		set:        make([]uint64, len(set)),
		reset:      make([]uint64, len(reset)),
		do:         onDo,
		availabler: avail,
	}

	for i, s := range set {
		if state, exists := m.statesMap[s.GetStateId()]; exists {
			a.set[i] = 1 << state.bit
		} else {
			return fmt.Errorf("state '%s' doesn't exists", s.GetStateId())
		}
	}

	for i, s := range reset {
		if state, exists := m.statesMap[s.GetStateId()]; exists {
			a.reset[i] = ^(1 << state.bit)
		} else {
			return fmt.Errorf("state '%s' doesn't exists", s.GetStateId())
		}
	}

	m.actionsMap[id] = a

	return nil
}

func (m *Multistate) MustAddAction(id, caption string, from expr.Expression, set, reset []State, onDo ActionDoFunc, avail Availabler) {
	if err := m.AddAction(id, caption, from, set, reset, onDo, avail); err != nil {
		panic(err)
	}
}

func (m *Multistate) AddCluster(name string, expr expr.Expression) {
	m.clusters = append(m.clusters, cluster{
		id:         uint8(len(m.clusters)),
		name:       name,
		expression: expr,
	})
}

func (m *Multistate) Compile() error {
	if m.statesActions != nil {
		return fmt.Errorf("multistate is already compiled")
	}

	m.statesActions = make(map[uint64]map[string]uint64)
	m.statesActions[0] = make(map[string]uint64)

	changed := true
	for changed {
		// println()
		changed = false

		for _, action := range m.actionsMap {
			for state, actions := range m.statesActions {
				if action.from.Eval(state) {
					if _, exists := actions[action.id]; !exists {
						changed = true
						newState := state
						for _, v := range action.set {
							// println("set", v)
							newState |= v
						}
						for _, v := range action.reset {
							newState &= v
							// println("reset", v)
						}

						// println("from", state, "to", newState, "by", action.id)

						actions[action.id] = newState
						if _, exists := m.statesActions[newState]; !exists {
							m.statesActions[newState] = make(map[string]uint64)
						}
					}
				}
			}
		}
	}

	m.stateClusterMap = map[uint64]*cluster{}
	for i, cluster := range m.clusters {
		for state := range m.statesActions {
			if !cluster.expression.Eval(state) {
				continue
			}
			if c, exists := m.stateClusterMap[state]; exists {
				return fmt.Errorf("the state %d exists at least in 2 clusters: %s and %s", state, c.name, cluster.name)
			}
			m.stateClusterMap[state] = &m.clusters[i]
		}
	}

	return nil
}

func (m *Multistate) MustCompile() {
	if err := m.Compile(); err != nil {
		panic(err)
	}
}

func (m *Multistate) GetStateActions(ctx context.Context, state uint64) []string {
	if actions, exists := m.statesActions[state]; exists {
		res := make([]string, 0, len(actions))

		for actionId := range actions {
			action := m.actionsMap[actionId]
			if action.availabler == nil || action.availabler.IsAvailable(ctx) {
				res = append(res, actionId)
			}
		}

		return res
	}

	return nil
}

func (m *Multistate) DoAction(ctx context.Context, entity Entity, action string, opts ...interface{}) (uint64, error) {
	ctx, err := entity.StartAction(ctx)
	if err != nil {
		return 0, entity.EndAction(ctx, err)
	}

	curState, err := entity.GetState(ctx)
	if err != nil {
		return 0, entity.EndAction(ctx, err)
	}

	actions, exists := m.statesActions[curState]
	if !exists {
		return 0, entity.EndAction(ctx, fmt.Errorf("current state %d: %w", curState, ErrInvalidState))
	}

	newState, exists := actions[action]
	if !exists {
		return 0, entity.EndAction(ctx, fmt.Errorf("action '%s', current state %d: %w", action, curState, ErrInvalidAction))
	}

	if avail := m.actionsMap[action].availabler; avail != nil && !avail.IsAvailable(ctx) {
		return 0, entity.EndAction(ctx, fmt.Errorf("action '%s', current state %d: %w", action, curState, ErrNotAvailable))
	}

	if m.onDo != nil {
		if err := m.onDo(ctx, entity, curState, newState, action, opts...); err != nil {
			return 0, entity.EndAction(ctx, fmt.Errorf("%w: %w", ErrExecutionAction, err))
		}
	}

	if onAction := m.actionsMap[action].do; onAction != nil {
		if err := onAction(ctx, entity, opts...); err != nil {
			return 0, entity.EndAction(ctx, fmt.Errorf("%w: %w", ErrExecutionAction, err))
		}
	}

	if err := entity.SetState(ctx, newState); err != nil {
		return 0, entity.EndAction(ctx, fmt.Errorf("%w: %w", ErrSetState, err))
	}

	return newState, entity.EndAction(ctx, nil)
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

func (m *Multistate) GetStatesByActions(actions ...string) []uint64 {
	set := make(map[uint64]struct{})

	for _, action := range actions {
		for state, mapact := range m.statesActions {
			if _, exists := mapact[action]; exists {
				set[state] = struct{}{}
			}
		}
	}

	ret := make([]uint64, 0, len(set))
	for el := range set {
		ret = append(ret, el)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })

	return ret
}

func (m *Multistate) GetMultistatesByStateIds(stateIds ...string) []uint64 {
	var bitmask uint64

	for _, id := range stateIds {
		if st, ok := m.statesMap[id]; ok {
			bitmask |= 1 << st.bit
		}
	}

	var ret []uint64

	for multistate := range m.statesActions {
		if multistate&bitmask != 0 {
			ret = append(ret, multistate)
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })

	return ret
}
