package main

import (
	. "./src/defs/"
	. "./src/driver/elevatorControls"
	. "./src/driver/io"
	//. "./src/network/localip"
	. "./src/network/netFwd"
	. "./src/network/peers"
	. "./src/orderLogic/elevatorManagement"
	. "fmt"
	"strconv"
	"time"
)

func main() {

	// Get preferred outbound ip of this machine
	myIP := GetOutboundIP()
	Println("My IP adress is", myIP)
	myID := IPToID[myIP]
	Println("My ID is", myID)
	state := Init

	sucess := IoInit()
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
		peerUpdate := make(chan PeerStatus)
		masterIDUpdate := make(chan int)
		masterBroadcast := make(chan PeerUpdate)
		masterBroadcastEnable := make(chan bool)

		movementInstructions := make(chan ElevatorMovement)
		statusReports := make(chan ElevatorStatus)
		statusReportsSend1 := make(chan ElevatorStatus)
		statusReportsSend2 := make(chan ElevatorStatus)
		movementReport := make(chan ElevatorMovement)

		buttonReports := make(chan int)
		buttonNewSend := make(chan int)
		buttonCompletedSend := make(chan int)
		orderQueueReport := make(chan OrderQueue)
		stateUpdate := make(chan int)

		statusMessage := make(chan StatusMessage)
		buttonNewRecieve := make(chan ButtonMessage)
		buttonCompletedRecieve := make(chan ButtonMessage)
		orderMessage := make(chan OrderMessage)
		orderMessageSend1 := make(chan OrderMessage)
		orderMessageSend2 := make(chan OrderMessage)
		orderMessageSend3 := make(chan OrderMessage)
		confirmedQueue := make(chan map[int]bool)

		go Receiver(12038, peerUpdateCh)
		go Transmitter(12038, strconv.Itoa(myID), peerTxEnable)
		go Receiver(11038, masterBroadcast)
		go Transmitter(11038, strconv.Itoa(myID), masterBroadcastEnable)

		go LocalElevator(movementInstructions, statusReports, movementReport)
		go MonitorOrderbuttons(buttonReports)

		go SendToNetwork(myID, masterIDUpdate, statusReportsSend1, buttonNewSend, buttonCompletedSend, orderQueueReport)
		go RecieveFromNetwork(myID, statusMessage, buttonNewRecieve, buttonCompletedRecieve, orderMessage)

		go CreateOrderQueue(stateUpdate, peerUpdate, statusMessage, buttonCompletedRecieve, buttonNewRecieve, orderQueueReport, orderMessageSend3)
		go Destination(statusReportsSend2, orderMessageSend1, movementInstructions)
		go BroadcastElevatorStatus(statusReports, statusReportsSend1, statusReportsSend2)
		go BroadcastOrderMessage(orderMessage, orderMessageSend1, orderMessageSend2, orderMessageSend3)
		go WatchCompletedOrders(movementReport, buttonCompletedSend)
		go WatchIncommingOrders(buttonReports, confirmedQueue, buttonNewSend)
		go CreateCurrentQueue(orderMessageSend2, confirmedQueue)

		establishConnection(peerUpdateCh, peerTxEnable, masterIDUpdate,
			masterBroadcast, masterBroadcastEnable, myID, &state)

		/*Println("Number of peers were", numberOfPeers)
		masterID := -1
		masterBroadcast := make(chan PeerUpdate)
		masterBroadcastEnable := make(chan bool)
		go Receiver(11038, masterBroadcast)
		go Transmitter(11038, strconv.Itoa(myID), masterBroadcastEnable)
		if numberOfPeers == 0 {
			Println("I am master", myID)
			masterID = myID
			masterBroadcastEnable <- true
			state = Master
		} else {
			m := <-masterBroadcast
			Printf("Peer update:\n")
			Printf("  Peers:    %q\n", m.Peers)
			Printf("  New:      %q\n", m.New)
			Printf("  Lost:     %q\n", m.Lost)
			masterID, _ = strconv.Atoi(m.Peers[0])
			Println("I am not master, master is", masterID)
			masterBroadcastEnable <- false
			state = Slave
		}
		masterIDUpdate <- masterID*/

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
				Println("Switched state to master")
				stateUpdate <- state
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
								peerUpdate <- PeerStatus{i, true}
							}
						}
						for _, ID := range p.Lost {
							Println("Lost peer", ID)
							i, _ := strconv.Atoi(ID)
							if i > 0 {
								peerUpdate <- PeerStatus{i, false}
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
					}
				}
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
									masterBroadcastEnable <- true
								}
								Println("New master", newMaster)
								masterIDUpdate <- newMaster
							}
						}
					case p = <-peerUpdateCh:
						if p.New != "" {
							i, _ := strconv.Atoi(p.New)
							if i > 0 {
								Println("newID as int =", i)
								Println("Gained peer", p.New)
								peerUpdate <- PeerStatus{i, true}
							}
						}
						for _, ID := range p.Lost {
							Println("Lost peer", ID)
							i, _ := strconv.Atoi(ID)
							if i > 0 {
								Println("Lost peer", i)
								peerUpdate <- PeerStatus{i, false}
								if i == myID {
									println("Detected lost connection")
									state = NoNetwork
								}
							}
						}
					}
				}
			case NoNetwork:
				//Internal buttons skal fortsatt betjenes.
				println("Detected lost connection and switched state")
				stateUpdate <- state
				for state == NoNetwork {
					p := <-peerUpdateCh
					if p.New != "" {
						numberOfPeers := len(p.Peers)
						i, _ := strconv.Atoi(p.New)
						if i > 0 {
							Println("in NoNetwork newID as int =", i)
							peerUpdate <- PeerStatus{i, true}
							Println("Gained peer", p.New)
						}
						println("Deciding if we are master or not:", numberOfPeers)
						if numberOfPeers == 1 {
							state = Master
						} else {
							state = Slave
						}
					}
				}
				/*Println("Switching state from ", state)
				establishConnection(peerUpdateCh, peerTxEnable, masterIDUpdate,
					masterBroadcast, masterBroadcastEnable, myID, &state)
				Println("Switching state to ", state)*/

			case DeadElevator:
				//Ingen knapper kan betjenes
				//Late som død for nettverket?
				for state == DeadElevator {
					continue
				}
			}
		}
	}

	//elevatorIsAlive := IoInit()

	/*if elevatorIsAlive {
		state = DeadElevator
	}*/
}

func establishConnection(peerUpdateCh <-chan PeerUpdate, peerTxEnable chan<- bool, masterIDUpdate chan<- int,
	masterBroadcast <-chan PeerUpdate, masterBroadcastEnable chan<- bool, myID int, state *int) {
	timer := time.NewTimer(45 * time.Millisecond)
	numberOfPeers := 0
	select {
	case p := <-peerUpdateCh:
		numberOfPeers = len(p.Peers)
	case <-timer.C:
		break
	}
	peerTxEnable <- true

	Println("Number of peers were", numberOfPeers)
	masterID := -1
	if numberOfPeers == 0 {
		Println("I am master", myID)
		masterID = myID
		masterBroadcastEnable <- true
		*state = Master
	} else {
		m := <-masterBroadcast
		Printf("Peer update:\n")
		Printf("  Peers:    %q\n", m.Peers)
		Printf("  New:      %q\n", m.New)
		Printf("  Lost:     %q\n", m.Lost)
		masterID, _ = strconv.Atoi(m.Peers[0])
		Println("I am not master, master is", masterID)
		masterBroadcastEnable <- false
		*state = Slave
	}
	masterIDUpdate <- masterID
}
