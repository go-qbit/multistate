package multistate

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"

	"github.com/tmc/dot"
)

func (m *Multistate) GetGraphSVG() string {
	g := dot.NewGraph("Multistate")

	nodes := make(map[uint64]*dot.Node)

	for state := range m.statesActions {
		var strFlags string
		for i, flag := range m.GetStateFlags(state) {
			if i > 0 {
				strFlags += "<BR/>"
			}
			strFlags += fmt.Sprintf("<I>[% 2d]</I> %s", flag.Bit, flag.Caption)
		}
		if strFlags == "" {
			strFlags = m.emptyStateName
		}
		if strFlags == "" {
			strFlags = "EMPTY"
		}

		hs := md5.New()
		_ = binary.Write(hs, binary.LittleEndian, state)
		digestBuf := bytes.NewBuffer(hs.Sum(nil))
		var c1, c2 uint32
		_ = binary.Read(digestBuf, binary.LittleEndian, &c1)
		_ = binary.Read(digestBuf, binary.LittleEndian, &c2)

		color := fmt.Sprintf("%f %f %f", float64(c1)/float64(math.MaxUint32),
			float64(c2)/float64(math.MaxUint32),
			0.7)

		n := dot.NewNode(strconv.FormatUint(state, 16))
		_ = n.Set("shape", "plaintext")
		_ = n.Set("label", fmt.Sprintf(`<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0"><TR><TD><B>%d</B></TD><TD>%s</TD></TR></TABLE>>`, state, strFlags))
		_ = n.Set("color", color)
		_ = n.Set("fontcolor", color)
		g.AddNode(n)
		nodes[state] = n
	}

	for from, actions := range m.statesActions {
		for action, to := range actions {
			e := dot.NewEdge(nodes[from], nodes[to])
			if m.actionsMap[action].permission != nil {
				_ = e.Set("label", fmt.Sprintf("%s\n(%s)\n[%s]", m.actionsMap[action].caption, action,
					m.actionsMap[action].permission.GetGroupId()+"."+m.actionsMap[action].permission.GetId()))
			} else {
				_ = e.Set("label", fmt.Sprintf("%s\n(%s)", m.actionsMap[action].caption, action))
			}
			color := nodes[from].Get("color")
			_ = e.Set("color", color)
			_ = e.Set("fontcolor", color)
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
