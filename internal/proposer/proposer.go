package proposer

import (
	"fmt"
	"image/color"
	"log"
	. "paxos-simulator/internal"
	. "paxos-simulator/internal/network"
	"strconv"
	"sync/atomic"
	"time"

	g "github.com/AllenDang/giu"
)

type Proposer struct {
	GroupID        int
	NodeID         int
	Acceptors      []string
	Net            Network
	HighestID      int64
	ProposalInHand Proposal
	PrepareCount   atomic.Int32
	ProposerState  int
	input          inputModel
}

type inputModel struct {
	ProposalValue int32
}

func StateToString(state int) string {
	switch state {
	case 1:
		return "PREPARE"
	case 2:
		return "PROPOSE"
	case 3:
		return "FAILED"
	}
	return "IDLE"
}

func StateToColor(state int) color.RGBA {
	switch state {
	case 1:
		return color.RGBA{R: 128, G: 128, B: 255, A: 255}
	case 2:
		return color.RGBA{R: 128, G: 255, B: 128, A: 255}
	case 3:
		return color.RGBA{R: 255, G: 128, B: 128, A: 255}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

func NewProposer(groupID int, nodeID int, network Network) *Proposer {
	return &Proposer{
		GroupID: groupID,
		NodeID:  nodeID,
		Net:     network,
	}
}

func (p *Proposer) Name() string {
	return fmt.Sprintf("[%d]proposer%d", p.GroupID, p.NodeID)
}

func (p *Proposer) UpdateAcceptor(acceptors []string) {
	p.Acceptors = acceptors
}

func (p *Proposer) propose() {
	if p.ProposerState != 0 && p.ProposerState != 3 {
		p.log("Proposer is busy, current state: %s", StateToString(p.ProposerState))
		return
	}
	p.ProposalInHand.ProposalID = time.Now().UnixNano()
	p.ProposalInHand.Value = int(p.input.ProposalValue)
	p.PrepareCount.Store(0)
	// Prepare
	// need a timer to avoid timout
	ticker := time.NewTicker(10 * time.Second)
	currentID := p.ProposalInHand.ProposalID
	go func() {
		<-ticker.C
		if currentID == p.ProposalInHand.ProposalID && p.ProposerState != 0 {
			p.log("Prepare timeout")
			p.ProposerState = 3
			p.PrepareCount.Store(0)
			p.HighestID = 0
		}
	}()
	p.ProposerState = 1
	for _, acceptor := range p.Acceptors {
		p.Net.Send(acceptor, NetworkMessage{
			Sender: p.Name(),
			Data: MessagePrepare{
				ProposalID: p.ProposalInHand.ProposalID,
			},
		})
	}
}

func (p *Proposer) doPropose() {
	p.log("Get half acceptors's response, starting propose")
	p.PrepareCount.Store(0)
	p.ProposerState = 2
	for _, acceptor := range p.Acceptors {
		p.Net.Send(acceptor, NetworkMessage{
			Sender: p.Name(),
			Data: MessageProposal{
				ProposalID: p.ProposalInHand.ProposalID,
				Value:      p.ProposalInHand.Value,
			},
		})
	}
	p.ProposerState = 0
	p.HighestID = 0
}

func (p *Proposer) Run() {
	p.log("Starting")
	for {
		msg := p.Net.Receive(p.Name())
		if p.ProposerState == 2 {
			continue
		}
		switch msg.Data.(type) {
		case MessageProposal:
			// Only receive prepare response when in prepare state
			if p.ProposerState != 1 {
				continue
			}
			msg := msg.Data.(MessageProposal)
			if msg.ProposalID > p.HighestID {
				p.HighestID = msg.ProposalID
				p.ProposalInHand.Value = msg.Value
				p.log(fmt.Sprintf("Acceptor already accepted a value %v", msg.Value))
			}
			current := p.PrepareCount.Add(1)
			p.log("Received response of prepare")
			if int(current) > len(p.Acceptors)/2 {
				// Propose
				p.doPropose()
			}
		}
	}
}

func (p *Proposer) log(format string, v ...any) {
	log.Printf("[%s] %s", p.Name(), fmt.Sprintf(format, v...))
}

func (p *Proposer) Render(x float32, y float32) {
	g.Window(p.Name()).Pos(x, y).Size(180, 250).Layout(
		g.Label("HighestID: "+strconv.Itoa(int(p.HighestID))),
		g.Label("ProposalID: "+strconv.Itoa(int(p.ProposalInHand.ProposalID/1000%1e9))),
		g.Label("ProposalValue: "+strconv.Itoa(p.ProposalInHand.Value)),
		g.Label("PrepareCount: "+strconv.Itoa(int(p.PrepareCount.Load()))),
		g.Style().SetColor(g.StyleColorText, StateToColor(p.ProposerState)).To(
			g.Label("State: "+StateToString(p.ProposerState)),
		),
		g.Separator(),
		g.InputInt(&p.input.ProposalValue).Label("Value"),
		g.Button("Propose").OnClick(func() {
			go p.propose()
		}),
	)
}
