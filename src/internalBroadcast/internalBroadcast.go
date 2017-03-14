package internalBroadcast

import (
	. "fmt"
	. "../defs/"
)
//These functions takes values on a single channel and copies them to severl others.


func BroadcastElevatorStatus(statusReport <-chan ElevatorStatus, send1, send2, send3 chan<- ElevatorStatus) {
	for {
		select {
		case status := <-statusReport:
			Println("Sending 1")
			send1 <- status
			Println("Sending 2")
			send2 <- status
			Println("Sending 3")
			send3 <- status
			Println("sent all")
		}
	}
}

func BroadcastOrderMessage(orderMessage <-chan OrderMessage, send1, send2, send3 chan<- OrderMessage) {
	for {
		select {
		case order := <-orderMessage:
			send1 <- order
			send2 <- order
			send3 <- order
		}
	}
}

func BroadcastStateUpdates(stateUpdate <-chan int, send1, send2, send3 chan<- int) {
	for  {
		select {
		case status := <-stateUpdate:
			send1 <- status
			send2 <- status
			send3 <- status
		}
	}
}

func BroadcastPeerUpdates(PeerUpdate <-chan PeerStatus, send1, send2 chan<- PeerStatus) {
	for {
		select {
		case status := <-PeerUpdate:
			send1 <- status
			send2 <- status
		}
	}
}