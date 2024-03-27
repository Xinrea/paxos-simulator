package network

import (
	"math/rand"
	"sync"
	"time"

	g "github.com/AllenDang/giu"
)

const MessageBuffer = 128

type Network interface {
	Send(id string, msg NetworkMessage)
	Receive(id string) NetworkMessage
}

type NetworkMessage struct {
	Sender string
	Data   interface{}
}

type ChanNetwork struct {
	lock  sync.Mutex
	chans map[string]chan NetworkMessage
	loss  float32
}

func NewChanNetwork(loss float32) *ChanNetwork {
	return &ChanNetwork{
		chans: make(map[string]chan NetworkMessage),
		loss:  loss,
	}
}

func (n *ChanNetwork) Send(id string, message NetworkMessage) {
	if rand.Float64() < float64(n.loss) {
		return
	}
	n.lock.Lock()
	if _, ok := n.chans[id]; !ok {
		n.chans[id] = make(chan NetworkMessage, MessageBuffer)
	}
	n.lock.Unlock()
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	n.chans[id] <- message
}

func (n *ChanNetwork) Receive(id string) NetworkMessage {
	n.lock.Lock()
	if _, ok := n.chans[id]; !ok {
		n.chans[id] = make(chan NetworkMessage, MessageBuffer)
	}
	n.lock.Unlock()
	// add random delay for visualization
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	return <-n.chans[id]
}

func (n *ChanNetwork) Render(x, y float32) {
	g.Window("Network").Pos(x, y).Size(300, 80).Layout(
		g.SliderFloat(&n.loss, 0, 1).Label("Loss Rate"),
	)
}
