package network

// Message for proposer->acceptor or acceptor->proposer (prepare)
type MessageProposal struct {
	ProposalID int64
	Value      int
}

// Message for proposer->acceptor
type MessagePrepare struct {
	ProposalID int64
}

// Message for acceptor->learner
type MessageLearn struct {
	Value int
}
