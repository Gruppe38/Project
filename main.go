package main

import (
	. "./src/defs/"
	. "./src/driver/elevatorControls"
	. "./src/driver/io"
	. "./src/network/localip"
	. "./src/network/peers"
	//"time"
	. "fmt"
)

func main() {
	var state = Init

	Println(IoInit())
	//SimInit()

	movementInstructions := make(chan ElevatorMovement)
	statusReports := make(chan ElevatorStatus)
	shutdownElevator := make(chan bool)
	go LocalElevator(movementInstructions, statusReports, shutdownElevator)

	buttonReports := make(chan int)
	shutdownMonitor := make(chan bool)
	go MonitorOrderbuttons(buttonReports, shutdownMonitor)

	/*for {
		select {
		case v := <-buttonReports:
			Println(v)
		case v := <-statusReports:
			Println(v)
		}
	}*/
	movementInstructions <- ElevatorMovement{false, 3}
	for {
		switch state {
		case Init:
			continue
		case Master:
			continue
		case Slave:
			continue
		case NoNetwork:
			continue
		case DeadElevator:
			continue
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
