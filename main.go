package main

import (
	. "./src/defs/"
	. "./src/driver/elevatorControls"
	. "./src/driver/io"
	. "./src/internalBroadcast"
	. "./src/network/netFwd"
	. "./src/network/peers"
	. "./src/orderLogic/elevatorManagement"
	. "./src/orderLogic/orders"
	. "./src/stateMachine/"
	. "fmt"
	"strconv"
)

func main() {

	// Get preferred outbound ip of this machine
	myIP := GetOutboundIP()
	Println("My IP adress is", myIP)
	myID := IPToID[myIP]
	Println("My ID is", myID)
	state := Init

	sucess := IoInit()
	Println("Init sucess: ", sucess)

	if sucess {
		peerUpdateCh := make(chan PeerUpdate)
		peerTxEnable := make(chan bool)
		peerStatusUpdate := make(chan PeerStatus)
		masterBroadcast := make(chan PeerUpdate)
		masterBroadcastEnable := make(chan bool)
		peerChannels := PeerChannels{peerUpdateCh, peerTxEnable, peerStatusUpdate, masterBroadcast, masterBroadcastEnable}

		peerStatusUpdateSend1 := make(chan PeerStatus)
		peerStatusUpdateSend2 := make(chan PeerStatus)
		masterIDUpdate := make(chan int)
		masterIDUpdateSend1 := make(chan int)
		masterIDUpdateSend2 := make(chan int)

		statusReportsSend1 := make(chan ElevatorStatus)
		buttonNewSend := make(chan int)
		buttonCompletedSend := make(chan int)
		orderQueueReport := make(chan OrderQueue)
		sendChannels := SendChannels{statusReportsSend1, buttonNewSend, buttonCompletedSend, orderQueueReport}

		statusMessage := make(chan StatusMessage, 99)
		buttonNewRecieve := make(chan ButtonMessage, 99)
		buttonCompletedRecieve := make(chan ButtonMessage, 99)
		orderMessage := make(chan OrderMessage, 99)
		recieveChannels := RecieveChannels{statusMessage, buttonNewRecieve, buttonCompletedRecieve, orderMessage}

		movementInstructions := make(chan ElevatorMovement)
		statusReports := make(chan ElevatorStatus)
		statusReportsSend2 := make(chan ElevatorStatus)
		statusReportsSend3 := make(chan ElevatorStatus)
		movementReport := make(chan ElevatorMovement)

		stateUpdate := make(chan int)
		stateUpdateSend1 := make(chan int)
		stateUpdateSend2 := make(chan int)
		stateUpdateSend3 := make(chan int)
		pushOrdersToMaster := make(chan bool)

		orderMessageSend1 := make(chan OrderMessage)
		orderMessageSend2 := make(chan OrderMessage)
		orderMessageSend3 := make(chan OrderMessage)
		confirmedQueue := make(chan map[int]bool)

		go Receiver(12038, peerUpdateCh)
		go Transmitter(12038, strconv.Itoa(myID), peerTxEnable)
		go Receiver(11038, masterBroadcast)
		go Transmitter(11038, strconv.Itoa(myID), masterBroadcastEnable)

		go BroadcastStateUpdates(stateUpdate, stateUpdateSend1, stateUpdateSend2, stateUpdateSend3)
		go BroadcastPeerUpdates(peerStatusUpdate, peerStatusUpdateSend1, peerStatusUpdateSend2)
		go BroadcastElevatorStatus(statusReports, statusReportsSend1, statusReportsSend2, statusReportsSend3)
		go BroadcastOrderMessage(orderMessage, orderMessageSend1, orderMessageSend2, orderMessageSend3)
		go BroadcastMasterUpdate(masterIDUpdate, masterIDUpdateSend1, masterIDUpdateSend2)

		go ExecuteInstructions(movementInstructions, statusReports, movementReport)

		go SendToNetwork(myID, masterIDUpdateSend1, peerStatusUpdateSend2, stateUpdateSend2, sendChannels)
		go RecieveFromNetwork(myID, masterIDUpdateSend2, stateUpdateSend3, recieveChannels)

		go CreateOrderQueue(stateUpdateSend1, peerStatusUpdateSend1, statusMessage, buttonCompletedRecieve, buttonNewRecieve, orderQueueReport, orderMessageSend3)
		go AssignMovementInstruction(statusReportsSend2, orderMessageSend1, movementInstructions)

		go WatchCompletedOrders(movementReport, buttonCompletedSend)
		go WatchIncommingOrders(confirmedQueue, buttonNewSend, pushOrdersToMaster)
		go CreateCurrentQueue(orderMessageSend2, confirmedQueue)

		EstablishConnection(peerUpdateCh, peerTxEnable, masterIDUpdate,
			masterBroadcast, masterBroadcastEnable, myID, &state)

		RunElevator(state, myID, stateUpdate, statusReportsSend3, masterIDUpdate, pushOrdersToMaster, peerChannels, sendChannels, recieveChannels)
	}
}
