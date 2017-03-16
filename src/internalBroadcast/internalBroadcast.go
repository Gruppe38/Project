package internalBroadcast

import (
	. "../defs/"
)

//These functions takes values on a single channel and copies them to severl others.

func BroadcastElevatorStatus(statusReport <-chan ElevatorStatus, send1, send2, send3 chan<- ElevatorStatus) {
	for {
		select {
		case status := <-statusReport:
			send1 <- status
			send2 <- status
			send3 <- status
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
	for {
		select {
		case status := <-stateUpdate:
			send1 <- status
			send2 <- status
			send3 <- status
		}
	}
}

func BroadcastPeerUpdates(PeerUpdate <-chan PeerStatus, send1, send2, send3 chan<- PeerStatus) {
	for {
		select {
		case status := <-PeerUpdate:
			send1 <- status
			send2 <- status
			send3 <- status
		}
	}
}

func BroadcastMasterUpdate(updateMaster <-chan int, send1, send2 chan<- int) {
	for {
		select {
		case master := <-updateMaster:
			send1 <- master
			send2 <- master
		}
	}
}
