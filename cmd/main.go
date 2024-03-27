package main

import (
	. "paxos-simulator/internal/acceptor"
	. "paxos-simulator/internal/learner"
	. "paxos-simulator/internal/network"
	. "paxos-simulator/internal/proposer"
	"time"

	g "github.com/AllenDang/giu"
)

func main() {
	// setup internal network
	network := NewChanNetwork(0)
	// setup learners
	learnerNum := 3
	learners := make([]*Learner, learnerNum)
	learnersList := []string{}
	for i := range learnerNum {
		learners[i] = NewLearner(0, i, network)
		learnersList = append(learnersList, learners[i].Name())
		go learners[i].Run()
	}
	// setup acceptors
	acceptorNum := 5
	acceptors := make([]*Acceptor, acceptorNum)
	for i := range acceptorNum {
		acceptors[i] = NewAcceptor(network, 0, i)
		acceptors[i].UpdateLearners(learnersList)
		go acceptors[i].Run()
	}
	acceptorList := []string{}
	for _, a := range acceptors {
		acceptorList = append(acceptorList, a.Name())
	}
	// setup proposers
	proposerNum := 3
	proposers := make([]*Proposer, proposerNum)
	for i := range proposerNum {
		proposers[i] = NewProposer(0, i, network)
		proposers[i].UpdateAcceptor(acceptorList)
		go proposers[i].Run()
	}
	// visualize
	loop := func() {
		g.MainMenuBar().Layout(g.Menu("New").Layout(
			g.MenuItem("New Proposer").OnClick(func() {
				proposers = append(proposers, NewProposer(0, len(proposers), network))
				proposers[len(proposers)-1].UpdateAcceptor(acceptorList)
				go proposers[len(proposers)-1].Run()
			}),
			g.MenuItem("New Acceptor").OnClick(func() {
				acceptors = append(acceptors, NewAcceptor(network, 0, len(acceptors)))
				acceptorList = append(acceptorList, acceptors[len(acceptors)-1].Name())
				for _, p := range proposers {
					p.UpdateAcceptor(acceptorList)
				}
				go acceptors[len(acceptors)-1].Run()
			}),
			g.MenuItem("New Learner").OnClick(func() {
				learners = append(learners, NewLearner(0, len(learners), network))
				learnersList = append(learnersList, learners[len(learners)-1].Name())
				for _, a := range acceptors {
					a.UpdateLearners(learnersList)
				}
				go learners[len(learners)-1].Run()
			}),
		)).Build()
		network.Render(10, 700)
		for i, p := range proposers {
			p.Render(10+float32(i)*200, 30)
		}
		for i, a := range acceptors {
			a.Render(10+float32(i)*200, 350)
		}
		for i, l := range learners {
			l.Render(10+float32(i)*170, 550)
		}
	}
	wnd := g.NewMasterWindow("paxos-simulator", 1024, 1024, g.MasterWindowFlagsFloating)
	go refreshUI()
	wnd.Run(loop)
}

func refreshUI() {
	for {
		ticker := time.NewTicker(100 * time.Millisecond)
		g.Update()
		<-ticker.C
	}
}
