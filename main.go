package main

import (
	. "./src/defs/"
	. "./src/driver/elevatorControls"
	. "./src/driver/io"
	//. "./src/network/localip"
	. "./src/network/netFwd"
	. "./src/network/peers"
	. "./src/orderLogic/elevatorManagement"
	//. "fmt"
	"time"
)

func main() {
	var state = Init

	sucess := IoInit()
	//SimInit()
	//sucess := true

	if sucess {
		peerUpdateCh := make(chan PeerUpdate)
		peerTxEnable := make(chan bool)

		go Receiver(12038, peerUpdateCh)
		//id := GetProcessID()

		timer := time.NewTimer(45 * time.Millisecond)
		numberOfPeers := 0
		select {
		case p := <-peerUpdateCh:
			numberOfPeers = len(p.Peers)
		case <-timer.C:
			break
		}
		id := string(numberOfPeers + 1)
		go Transmitter(12038, id, peerTxEnable)

		movementInstructions := make(chan ElevatorMovement)
		statusReports := make(chan ElevatorStatus)
		statusReportsSend1 := make(chan ElevatorStatus)
		statusReportsSend2 := make(chan ElevatorStatus)
		statusReportsSend3 := make(chan ElevatorStatus)

		buttonReports := make(chan int)
		myIDSend := make(chan int)
		myIDRecieve := make(chan int)
		masterID := make(chan int)
		buttonNewSend := make(chan int)
		buttonCompletedSend := make(chan int)
		orderQueueReport := make(chan OrderQueue)

		statusMessage := make(chan StatusMessage)
		buttonNewRecieve := make(chan ButtonMessage)
		buttonCompletedRecieve := make(chan ButtonMessage)
		orderMessage := make(chan OrderMessage)
		orderMessageSend1 := make(chan OrderMessage)
		orderMessageSend2 := make(chan OrderMessage)
		confirmedQueue := make(chan map[int]bool)

		go LocalElevator(movementInstructions, statusReports)
		go MonitorOrderbuttons(buttonReports)
		go CreateOrderQueue(statusMessage, buttonCompletedRecieve, buttonNewRecieve, orderQueueReport)

		go SendToNetwork(myIDSend, masterID, statusReportsSend1, buttonNewSend, buttonCompletedSend, orderQueueReport)
		go RecieveFromNetwork(myIDRecieve, statusMessage, buttonNewRecieve, buttonCompletedRecieve, orderMessage)

		go Destination(statusReportsSend2, orderMessageSend1, movementInstructions)
		go BroadcastElevatorStatus(statusReports, statusReportsSend1, statusReportsSend2, statusReportsSend3)
		go BroadcastOrderMessage(orderMessage, orderMessageSend1, orderMessageSend2)
		go WatchCompletedOrders(statusReportsSend3, buttonCompletedSend)
		go WatchIncommingOrders(buttonReports, confirmedQueue, buttonNewSend)
		go CreateCurrentQueue(orderMessageSend2, confirmedQueue)

		time.Sleep(500*time.Millisecond)
		myIDSend <- 1
		myIDRecieve <- 1
		masterID <- 1
		/*for {
			select {
			case v := <-buttonReports:
				Println(v)
			case v := <-statusReports:
				Println(v)
			}
		}*/
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
	}

	//elevatorIsAlive := IoInit()

	/*if elevatorIsAlive {
		state = DeadElevator
	}*/
}
