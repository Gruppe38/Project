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
	var state = Init

	sucess := IoInit()
	//SimInit()
	//sucess := true

	if sucess {
		peerUpdateCh := make(chan PeerUpdate)
		peerTxEnable := make(chan bool)

		go Receiver(12038, peerUpdateCh)
		go Transmitter(12038, strconv.Itoa(myID), peerTxEnable)

		timer := time.NewTimer(45 * time.Millisecond)
		numberOfPeers := 0
		select {
		case p := <-peerUpdateCh:
			numberOfPeers = len(p.Peers)
		case <-timer.C:
			break
		}
		peerTxEnable <- true

		movementInstructions := make(chan ElevatorMovement)
		statusReports := make(chan ElevatorStatus)
		statusReportsSend1 := make(chan ElevatorStatus)
		statusReportsSend2 := make(chan ElevatorStatus)
		movementReport := make(chan ElevatorMovement)

		buttonReports := make(chan int)
		masterIDUpdate := make(chan int)
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

		peerUpdate := make(chan PeerStatus)

		go LocalElevator(movementInstructions, statusReports, movementReport)
		go MonitorOrderbuttons(buttonReports)
		go CreateOrderQueue(stateUpdate, peerUpdate, statusMessage, buttonCompletedRecieve, buttonNewRecieve, orderQueueReport, orderMessageSend3)

		go SendToNetwork(myID, masterIDUpdate, statusReportsSend1, buttonNewSend, buttonCompletedSend, orderQueueReport)
		go RecieveFromNetwork(myID, statusMessage, buttonNewRecieve, buttonCompletedRecieve, orderMessage)

		go Destination(statusReportsSend2, orderMessageSend1, movementInstructions)
		go BroadcastElevatorStatus(statusReports, statusReportsSend1, statusReportsSend2)
		go BroadcastOrderMessage(orderMessage, orderMessageSend1, orderMessageSend2, orderMessageSend3)
		go WatchCompletedOrders(movementReport, buttonCompletedSend)
		go WatchIncommingOrders(buttonReports, confirmedQueue, buttonNewSend)
		go CreateCurrentQueue(orderMessageSend2, confirmedQueue)

		Println("Number of peers were", numberOfPeers)
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
		masterIDUpdate <- masterID

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
							}
						}
					}
				}
			case NoNetwork:
				//Internal buttons skal fortsatt betjenes.
				for state == NoNetwork {
					continue
				}
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
