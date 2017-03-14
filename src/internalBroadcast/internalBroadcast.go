package internalBroadcast

import (
	. "fmt"
	. "../defs/"
)

//statusReport inneholder heistatus, kommer fra watchElevator, som får kannalen via LocalElevator
//Send 1 til createOrderQueue, via nettverk
//send2 til destination
func BroadcastElevatorStatus(statusReport <-chan ElevatorStatus, send1, send2, send3 chan<- ElevatorStatus) {
	quit := false
	for !quit {
		select {
		case status, t := <-statusReport:
			if t {
				Println("Sending 1")
				send1 <- status
				Println("Sending 2")
				send2 <- status
				Println("Sending 3")
				send3 <- status
				Println("sent all")
			} else {
				close(send1)
				close(send2)
				close(send3)
				quit = true
			}
		}
	}
}

func BroadcastOrderMessage(orderMessage <-chan OrderMessage, send1, send2, send3 chan<- OrderMessage) {
	quit := false
	for !quit {
		select {
		case order, t := <-orderMessage:
			if t {
				send1 <- order
				send2 <- order
				send3 <- order
			} else {
				close(send1)
				close(send2)
				quit = true
			}
		}
	}
}

func BroadcastStateUpdates(stateUpdate <-chan int, send1, send2, send3 chan<- int) {
	quit := false
	for !quit {
		select {
		case status, t := <-stateUpdate:
			if t {
				send1 <- status
				send2 <- status
				send3 <- status
			} else {
				close(send1)
				close(send2)
				close(send3)
				quit = true
			}
		}
	}
}

func BroadcastPeerUpdates(PeerUpdate <-chan PeerStatus, send1, send2 chan<- PeerStatus) {
	quit := false
	for !quit {
		select {
		case status, t := <-PeerUpdate:
			if t {
				send1 <- status
				send2 <- status
			} else {
				close(send1)
				close(send2)
				quit = true
			}
		}
	}
}