package acceptor

import (
	"fmt"
	"log"
	. "paxos-simulator/internal"
	. "paxos-simulator/internal/network"
	"strconv"

	g "github.com/AllenDang/giu"
)

type Acceptor struct {
	GroupID          int
	NodeID           int
	HighestID        int64
	Accepted         bool
	AcceptedProposal Proposal
	Net              Network
	Disabled         bool
	Learners         []string
}

func NewAcceptor(net Network, groupID int, nodeID int) *Acceptor {
	return &Acceptor{
		GroupID: groupID,
		NodeID:  nodeID,
		Net:     net,
	}
}

func (a *Acceptor) Name() string {
	return fmt.Sprintf("[%d]acceptor%d", a.GroupID, a.NodeID)
}

func (a *Acceptor) UpdateLearners(learners []string) {
	a.Learners = learners
}

func (a *Acceptor) Run() {
	for {
		msg := a.Net.Receive(a.Name())
		if a.Disabled {
			continue
		}
		switch msg.Data.(type) {
		case MessagePrepare:
			sender := msg.Sender
			msg := msg.Data.(MessagePrepare)
			a.log("Received a prepare request with id: %d", msg.ProposalID)
			if msg.ProposalID < a.HighestID {
				// ignore
				continue
			}
			a.HighestID = msg.ProposalID
			// accepted_id <= highest_id
			// so here msg.ProposalID > highest_id >= accepted_id
			if a.Accepted {
				a.log("Already accepted a proposal: %v", a.AcceptedProposal)
				// reply with accepted value
				a.Net.Send(sender, NetworkMessage{
					Sender: a.Name(),
					Data: MessageProposal{
						ProposalID: a.AcceptedProposal.ProposalID,
						Value:      a.AcceptedProposal.Value,
					},
				})
			} else {
				// reply with none
				a.Net.Send(sender, NetworkMessage{
					Sender: a.Name(),
					Data: MessageProposal{
						ProposalID: 0,
					},
				})
			}
		case MessageProposal:
			msg := msg.Data.(MessageProposal)
			if msg.ProposalID >= a.HighestID {
				a.log("Accept proposal: %v", msg)
				a.HighestID = msg.ProposalID
				a.Accepted = true
				a.AcceptedProposal.ProposalID = msg.ProposalID
				a.AcceptedProposal.Value = msg.Value
				// send to all learners
				for _, l := range a.Learners {
					a.Net.Send(l, NetworkMessage{
						Sender: a.Name(),
						Data: MessageProposal{
							ProposalID: msg.ProposalID,
							Value:      msg.Value,
						},
					})
				}
			} else {
				a.log("Received a invalid proposal, %d < %d", msg.ProposalID, a.HighestID)
			}
		}
	}
}

func (a *Acceptor) log(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[%s] %s", a.Name(), msg)
}

func (a *Acceptor) Render(x, y float32) {
	g.Window(a.Name()).Pos(x, y).Size(180, 150).Layout(
		g.Label("HighestID: "+strconv.Itoa(int(a.HighestID/1000%1e9))),
		g.Label("Accepted: "+strconv.FormatBool(a.Accepted)),
		g.Label("AcceptedID: "+strconv.Itoa(int(a.AcceptedProposal.ProposalID/1000%1e9))),
		g.Label("AcceptedValue: "+strconv.Itoa(a.AcceptedProposal.Value)),
		g.Separator(),
		g.Checkbox("Disabled", &a.Disabled),
	)
}
