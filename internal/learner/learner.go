package learner

import (
	"fmt"
	"image/color"
	. "paxos-simulator/internal/network"

	g "github.com/AllenDang/giu"
)

type Learner struct {
	GroupID int
	NodeID  int
	Net     Network
	Value   int
	Learned bool
}

func NewLearner(groupID int, nodeID int, network Network) *Learner {
	return &Learner{
		GroupID: groupID,
		NodeID:  nodeID,
		Net:     network,
	}
}

func (l *Learner) Name() string {
	return fmt.Sprintf("[%d]learner%d", l.GroupID, l.NodeID)
}

func (l *Learner) Run() {
	for {
		msg := l.Net.Receive(l.Name())
		switch msg.Data.(type) {
		case MessageProposal:
			msg := msg.Data.(MessageProposal)
			l.Value = msg.Value
			l.Learned = true
		}
	}
}

func (l *Learner) Render(x, y float32) {
	// render learner
	if l.Learned {
		g.Window(l.Name()).Pos(x, y).Size(150, 100).Layout(
			g.Label("Value: "+fmt.Sprintf("%d", l.Value)),
			g.Style().SetColor(g.StyleColorText, color.RGBA{R: 128, G: 255, B: 128, A: 255}).To(
				g.Label("Learned: "+fmt.Sprintf("%v", l.Learned)),
			),
		)
	} else {
		g.Window(l.Name()).Pos(x, y).Size(150, 100).Layout(
			g.Label("Value: "+fmt.Sprintf("%d", l.Value)),
			g.Label("Learned: "+fmt.Sprintf("%v", l.Learned)),
		)
	}
}
