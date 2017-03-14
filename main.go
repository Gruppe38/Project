package main

import (
	. "./src/defs/"
	. "./src/driver/elevatorControls"
	. "./src/driver/io"
	//. "./src/network/localip"
	. "./src/network/netFwd"
	. "./src/network/peers"
	. "./src/orderLogic/elevatorManagement"
	. "./src/orderLogic/orders"
	. "fmt"
	. "./src/internalBroadcast"
	. "./src/stateMachine/"
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
	//SimInit()
	//sucess := true

	if sucess {

		/*timer := time.NewTimer(45 * time.Millisecond)
		numberOfPeers := 0
		select {
		case p := <-peerUpdateCh:
			numberOfPeers = len(p.Peers)
		case <-timer.C:
			break
		}
		peerTxEnable <- true*/
		peerUpdateCh := make(chan PeerUpdate)
		peerTxEnable := make(chan bool)
		peerStatusUpdate := make(chan PeerStatus)
		masterBroadcast := make(chan PeerUpdate)
		masterBroadcastEnable := make(chan bool)
		peerChannels := PeerChannels{peerUpdateCh, peerTxEnable, peerStatusUpdate, masterBroadcast, masterBroadcastEnable}

		peerStatusUpdateSend1 := make(chan PeerStatus)
		peerStatusUpdateSend2 := make(chan PeerStatus)
		masterIDUpdate := make(chan int)

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

		buttonReports := make(chan int)
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

		go LocalElevator(movementInstructions, statusReports, movementReport)
		go MonitorOrderbuttons(buttonReports)

		go SendToNetwork(myID, masterIDUpdate, peerStatusUpdateSend2, stateUpdateSend2, sendChannels)
		go RecieveFromNetwork(myID, stateUpdateSend3, recieveChannels)

		go CreateOrderQueue(stateUpdateSend1, peerStatusUpdateSend1, statusMessage, buttonCompletedRecieve, buttonNewRecieve, orderQueueReport, orderMessageSend3)
		go Destination(statusReportsSend2, orderMessageSend1, movementInstructions)

		go WatchCompletedOrders(movementReport, buttonCompletedSend)
		go WatchIncommingOrders(buttonReports, confirmedQueue, buttonNewSend, pushOrdersToMaster)
		go CreateCurrentQueue(orderMessageSend2, confirmedQueue)

		EstablishConnection(peerUpdateCh, peerTxEnable, masterIDUpdate,
			masterBroadcast, masterBroadcastEnable, myID, &state)

		RunElevator(state, myID, stateUpdate, statusReportsSend3, masterIDUpdate, pushOrdersToMaster, peerChannels, sendChannels, recieveChannels)
		/*for {
			switch state {
			case Init:
				continue
			case Master:
				Println("Switched state to master")
				stateUpdate <- state
				masterBroadcastEnable <- true
				for state == Master {
					select {
					case p := <-peerUpdateCh:
						Printf("Peer update:\n")
						Printf("  Peers:    %q\n", p.Peers)
						Printf("  New:      %q\n", p.New)
						Printf("  Lost:     %q\n", p.Lost)
						if p.New != "" {
							i, _ := strconv.Atoi(p.New)
							if i > 0 {
								Println("newID as int =", i)
								Println("Gained peer", p.New)
								peerStatusUpdate <- PeerStatus{i, true}
							}
						}
						for _, ID := range p.Lost {
							Println("Lost peer", ID)
							i, _ := strconv.Atoi(ID)
							if i > 0 {
								peerStatusUpdate <- PeerStatus{i, false}
								Println("Lost peer", i)
								if i == myID {
									println("Detected lost connection")
									state = NoNetwork
								}
							}
						}
					case m := <-masterBroadcast:
						Printf("Master update while master:\n")
						Printf("  Masters:    %q\n", m.Peers)
						Printf("  New:      %q\n", m.New)
						Printf("  Lost:     %q\n", m.Lost)
					case status := <-statusReportsSend3:
						if status.Timeout {
							state = DeadElevator
						}
					}
				}
				masterBroadcastEnable <- false
			case Slave:
				Println("Switched state to slave")
				stateUpdate <- state
				p := PeerUpdate{}
				for state == Slave {
					select {
					case m := <-masterBroadcast:
						Printf("Master update while slave:\n")
						Printf("  Masters:    %q\n", m.Peers)
						Printf("  New:      %q\n", m.New)
						Printf("  Lost:     %q\n", m.Lost)
						for _, ID := range m.Lost {
							i, _ := strconv.Atoi(ID)
							if i > 0 {
								Println("Lost master", i)
								t := time.NewTimer(30 * time.Millisecond)
								Println("Wait time started")
								waiting := true
								for waiting {
									select {
									case p = <-peerUpdateCh:
										Println("Got peer update while waiting for new master")
										continue
									case <-t.C:
										Println("Wait time for new master passed")
										waiting = false
									}
								}
								Println("Assigning new master")
								newMaster, _ := strconv.Atoi(p.Peers[0])
								if newMaster == myID {
									Println("I am master", myID)
									state = Master
								}
								Println("New master from lost", newMaster)
								masterIDUpdate <- newMaster
							}
						}
						for _, ID := range m.New {
							masterID := int(ID)
							Println("New master from new", masterID)
							masterIDUpdate <- masterID
						}
					case p = <-peerUpdateCh:
						if p.New != "" {
							i, _ := strconv.Atoi(p.New)
							if i > 0 {
								Println("newID as int =", i)
								Println("Gained peer", p.New)
								peerStatusUpdate <- PeerStatus{i, true}
							}
						}
						for _, ID := range p.Lost {
							Println("Lost peer", ID)
							i, _ := strconv.Atoi(ID)
							if i > 0 {
								Println("Lost peer", i)
								peerStatusUpdate <- PeerStatus{i, false}
								if i == myID {
									println("Detected lost connection")
									state = NoNetwork
								}
							}
						}
					case status := <-statusReportsSend3:
						if status.Timeout {
							state = DeadElevator
						}
					}
				}
			case NoNetwork:
				//Internal buttons skal fortsatt betjenes.
				println("Detected lost connection and switched state to NoNetwork")
				stateUpdate <- state
				stateUpdate2 := make(chan int)
				peerStatusUpdate <- PeerStatus{myID, true}
				numberOfPeers := 0
				masterID := -1
				stateUpdateDelay := time.NewTimer(45 * time.Millisecond)
				stateUpdateDelay.Stop()
				go DirectTransfer(myID, stateUpdate2, sendChannels, recieveChannels)
				for state == NoNetwork {
					select {
					case p := <-peerUpdateCh:
						Println("main  i no NoNetwork got peer update")
						if numberOfPeers == 0 {
							Println("main  i no NoNetwork started stateUpdateDelay")
							stateUpdateDelay.Reset(100 * time.Millisecond)
						}
						numberOfPeers = len(p.Peers)
					case status := <-statusReportsSend3:
						Println("Main in state noNewtwork got status update")
						if status.Timeout {
							state = DeadElevator
						}
					case m := <-masterBroadcast:
						if len(m.Peers) != 0 {
							masterID, _ = strconv.Atoi(m.Peers[0])
						} else {
							masterID = -1
						}
					case <-stateUpdateDelay.C:
						println("Deciding if we are master or not:", masterID)
						if masterID == -1 {
							state = Master
							masterIDUpdate <- myID
						} else {
							state = Slave
							masterIDUpdate <- masterID
							Println("Starting complete push to master")
							pushOrdersToMaster <- true
							Println("Waiting for complete push to master")
							<-pushOrdersToMaster
							Println("Completed push to master")
						}
					}
				}
				stateUpdate2 <- state

			case DeadElevator:
				Println("Entering state deadElevator")
				//Ingen knapper kan betjenes
				//Late som død for nettverket?
				stateUpdate <- state
				peerTxEnable <- false
				masterBroadcastEnable <- false
				for state == DeadElevator {
					//WriteAnalog(MOTOR, 0)
					select {
					case status := <-statusReportsSend3:
						if !status.Timeout {
							state = Slave
						}
					case <-peerUpdateCh:
					}
				}
				peerTxEnable <- true
				p := <-peerUpdateCh
				numberOfPeers := len(p.Peers)
				println("Deciding if we are master or not after reverting from timout:", numberOfPeers)
				if numberOfPeers == 1 {
					state = Master
				} else {
					state = Slave
				}
			}
		}*/
	}
}