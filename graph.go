package multistate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/tmc/dot"
)

func (m *Multistate) GetGraphSVG() string {
	g := dot.NewGraph("Multistate")

	nodes := make(map[uint64]*dot.Node)

	for state := range m.statesActions {
		n := dot.NewNode(fmt.Sprintf("(%d) %s", state, m.GetStateName(state)))
		g.AddNode(n)
		nodes[state] = n
	}

	for from, actions := range m.statesActions {
		for action, to := range actions {
			e := dot.NewEdge(nodes[from], nodes[to])
			if m.actionsMap[action].permission != nil {
				e.Set("label", fmt.Sprintf("%s\n(%s)\n[%s]", m.actionsMap[action].caption, action,
					m.actionsMap[action].permission.GetGroupId()+"."+m.actionsMap[action].permission.GetId()))
			} else {
				e.Set("label", fmt.Sprintf("%s\n(%s)", m.actionsMap[action].caption, action))
			}
			g.AddEdge(e)
		}
	}

	outBuf := &bytes.Buffer{}

	cmd := exec.Command("/usr/bin/dot", "-Tsvg")
	cmd.Stdin = bytes.NewBuffer([]byte(g.String()))
	cmd.Stdout = outBuf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	return outBuf.String()
}
