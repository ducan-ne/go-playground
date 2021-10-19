package main

import (
	"bytes"
	"fmt"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"math"
	"strings"
	"time"

	"golang.org/x/net/context"
)

const hb = 1

type node struct {
	id        int
	store     *raft.MemoryStorage
	pstore    map[string]string
	ctx       context.Context
	ctxCancel context.CancelFunc
	cfg       *raft.Config
	node      raft.Node
	ticker    <-chan time.Time
	done      <-chan struct{}
}

func newNode(id int, peers []raft.Peer) *node {
	n := &node{}
	ctx, cancel := context.WithCancel(context.Background())
	n.ctx = ctx
	n.ctxCancel = cancel
	n.store = raft.NewMemoryStorage()
	n.cfg = &raft.Config{
		ID:              uint64(id),
		ElectionTick:    10 * hb,
		HeartbeatTick:   hb,
		Storage:         n.store,
		MaxSizePerMsg:   math.MaxUint16,
		MaxInflightMsgs: 256,
	}
	n.pstore = make(map[string]string)
	n.id = id
	n.node = raft.StartNode(n.cfg, peers)
	return n
}

func (n *node) run() {
	n.ticker = time.Tick(time.Second)
	for {
		select {
		case <-n.ticker:
			n.node.Tick()
		case rd := <-n.node.Ready():
			n.saveToStorage(rd, rd.Entries, rd.Snapshot)
			n.send(rd.Messages)
			if !raft.IsEmptySnap(rd.Snapshot) {
				n.processSnapshot(rd.Snapshot)
			}
			for _, entry := range rd.CommittedEntries {
				n.process(entry)
				if entry.Type == raftpb.EntryConfChange {
					var cc raftpb.ConfChange
					cc.Unmarshal(entry.Data)
					n.node.ApplyConfChange(cc)
				}
			}
			n.node.Advance()
		case <-n.done:
			return
		}
	}
}

func (n *node) saveToStorage(rd raft.Ready, entries []raftpb.Entry, snapshot raftpb.Snapshot) {
	fmt.Println("Storing entries, node =", n.id)
	n.store.Append(entries)

	if !raft.IsEmptyHardState(rd.HardState) {
		fmt.Println("Setting hard state, node =", n.id)
		n.store.SetHardState(rd.HardState)
	}

	if !raft.IsEmptySnap(snapshot) {
		fmt.Println("Applying snapshot, node =", n.id)
		n.store.ApplySnapshot(snapshot)
	}
}

func (n *node) send(messages []raftpb.Message) {
	fmt.Println("*** Messages from NODE ", n.id, "***")
	fmt.Println("Count:", len(messages))

	for _, m := range messages {
		fmt.Println(raft.DescribeMessage(m, nil))

		// send message to other node
		nodes[int(m.To)].receive(n.ctx, m)
	}

	fmt.Println("***************")
}

func (n *node) processSnapshot(snapshot raftpb.Snapshot) {
	fmt.Println("Applying snapshot on", n.id, ":", snapshot)
	n.store.ApplySnapshot(snapshot)
}

func (n *node) process(entry raftpb.Entry) {
	fmt.Println("processing entry on ", n.id, ":", entry)
	if entry.Type == raftpb.EntryNormal && entry.Data != nil {
		fmt.Println("normal message:", string(entry.Data))
		parts := bytes.SplitN(entry.Data, []byte(":"), 2)
		fmt.Println(string(parts[0]), " = ", string(parts[1]))
		n.pstore[string(parts[0])] = string(parts[1])
	}
}

func (n *node) receive(ctx context.Context, message raftpb.Message) {
	fmt.Println("Received message, node =", n.id)
	n.node.Step(ctx, message)
}

var (
	nodes = make(map[int]*node)
)

func main() {
	nodes[1] = newNode(1, []raft.Peer{{ID: 1}, {ID: 2}, {ID: 3}})
	nodes[1].node.Campaign(nodes[1].ctx)
	go nodes[1].run()

	nodes[2] = newNode(2, []raft.Peer{{ID: 1}, {ID: 2}, {ID: 3}})
	go nodes[2].run()

	nodes[3] = newNode(3, []raft.Peer{{ID: 1}, {ID: 2}, {ID: 3}})
	go nodes[3].run()

	// Wait for leader, is there a better way to do this
	for nodes[1].node.Status().Lead != 1 {
		time.Sleep(100 * time.Millisecond)
	}

	nodes[2].node.Propose(nodes[2].ctx, []byte("mykey:myvalue"))

	// Wait for proposed entry to be commited in cluster
	// Probably a better way to check this
	time.Sleep(100 * time.Millisecond)

	fmt.Println(strings.Repeat("#", 20))
	for i, node := range nodes {
		fmt.Println("Node", i)
		for k, v := range node.pstore {
			fmt.Println(k, " =  ", v)
		}
		fmt.Println("")
	}
	fmt.Println(strings.Repeat("#", 20))

}
