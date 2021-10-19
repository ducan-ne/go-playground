package main

import (
	"bytes"
	"clgt.io/go-playground/hashicorpraft/store"
	"clgt.io/go-playground/hashicorpraft/transport"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

// Command line defaults
const (
	DefaultHTTPAddr = "localhost:11000"
	DefaultRaftAddr = "localhost:12000"
)

// Command line parameters
var httpAddr string
var raftAddr string
var joinAddr string
var nodeID string

func init() {
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&nodeID, "id", "", "Node ID")
}

func main() {
	flag.Parse()

	log.Println(raftAddr)
	log.Println(httpAddr)

	s := store.New()
	s.RaftDir = fmt.Sprintf("./raft/nodes/%s", nodeID)

	if joinAddr != "" {
		s.RaftBind = raftAddr
	} else {
		s.RaftBind = raftAddr
	}

	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := transport.New(httpAddr, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	// If join was specified, make the join request.
	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Println("hraftd started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("hraftd exiting")

}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
