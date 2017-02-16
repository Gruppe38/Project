package main

import (
		. "./src/defs/"
	//. "./src/driver/"
	. "./src/network/localip"
	. "./src/network/peers"
	//	. "fmt"
)

func main() {
	var state = Init
	for {
		switch state {
			case Init:{
				continue
			}
			case Master:{
				continue

			}
			case Slave:{
				continue

			}
			case NoNetwork:{
				continue

			}
			case DeadElevator:{
				continue
			}
		}
	}


	//elevatorIsAlive := IoInit()

	/*if elevatorIsAlive {
		state = DeadElevator
	}*/

	id := GetProcessID()

	peerUpdateCh := make(chan PeerUpdate)
	peerTxEnable := make(chan bool)

	go Transmitter(12038, id, peerTxEnable)
	go Receiver(12038, peerUpdateCh)

}
