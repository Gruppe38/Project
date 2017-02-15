package main

import (
	"./src/defs/"
	"./src/network/bcast"
	"./src/network/localip"
	"./src/network/peers"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(12038, id, peerTxEnable)
	go peers.Receiver(12038, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan definitions.HelloMsg)
	helloRx := make(chan definitions.HelloMsg)
	helloRx2 := make(chan int)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(13038, helloTx)
	go bcast.Receiver(13038, helloRx)
	go bcast.Receiver(14038, helloRx2)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := definitions.HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Receiveda: %#v\n", a)
		case b := <-helloRx2:
			fmt.Printf("Receivedb: %#v\n", b)
		}
	}
}
