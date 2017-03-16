package stateMachine

import (
	. "../defs/"
	. "../network/netFwd"
	. "fmt"
	"strconv"
	"time"
)

func RunElevator(state int, myID int, stateUpdate chan int, statusReportsSend3 chan ElevatorStatus, masterIDUpdate chan int, pushOrdersToMaster chan bool, peerChannels PeerChannels, sendChannels SendChannels, recieveChannels RecieveChannels) {
	for {
		switch state {
		case Init:
			continue
		//When an elevator is in case Master it calculates a cost and decides which elevator should take an order.
		case Master:
			Println("Current state: master")
			stateUpdate <- state
			peerChannels.MasterBroadcastEnable <- true
			for state == Master {
				select {
				case p := <-peerChannels.PeerUpdateCh:
					if p.New != "" {
						i, _ := strconv.Atoi(p.New)
						if i > 0 {
							peerChannels.PeerStatusUpdate <- PeerStatus{i, true}
						}
					}
					for _, ID := range p.Lost {
						i, _ := strconv.Atoi(ID)
						if i > 0 {
							peerChannels.PeerStatusUpdate <- PeerStatus{i, false}
							if i == myID {
								state = NoNetwork
							}
						}
					}
				case <-peerChannels.MasterBroadcast:
				case status := <-statusReportsSend3:
					if status.Timeout {
						state = DeadElevator
					}
				}
			}
			peerChannels.MasterBroadcastEnable <- false
		//When an elevator is in case Slave it recieves and order queue from master and decides which order should be handles first.
		case Slave:
			Println("Current state: slave")
			stateUpdate <- state
			p := PeerUpdate{}
			for state == Slave {
				select {
				case m := <-peerChannels.MasterBroadcast:
					for _, ID := range m.Lost {
						i, _ := strconv.Atoi(ID)
						if i > 0 {
							t := time.NewTimer(30 * time.Millisecond)
							waiting := true
							for waiting {
								select {
								case p = <-peerChannels.PeerUpdateCh:
									continue
								case <-t.C:
									waiting = false
								}
							}
							newMaster, _ := strconv.Atoi(p.Peers[0])
							Println("Peers list1: ", p.Peers)
							if newMaster == myID {
								state = Master
							}
							Println("New master1: ", newMaster)
							masterIDUpdate <- newMaster
						}
					}
					/*for _, ID := range m.New {
					masterID := int(ID)
					Println("m.new: ", m.New)
					Println("Peers list2: ", p.Peers)
					Println("New master2: ", masterID)
					masterIDUpdate <- masterID*/
					if m.New != "" {
						newMaster, _ := strconv.Atoi(m.New)
						masterIDUpdate <- newMaster
						Println("New master: ", newMaster)
					}
				case p = <-peerChannels.PeerUpdateCh:
					if p.New != "" {
						i, _ := strconv.Atoi(p.New)
						if i > 0 {
							peerChannels.PeerStatusUpdate <- PeerStatus{i, true}
						}
					}
					for _, ID := range p.Lost {
						i, _ := strconv.Atoi(ID)
						if i > 0 {
							peerChannels.PeerStatusUpdate <- PeerStatus{i, false}
							if i == myID {
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
		//When an elevator is in case NoElevator it no longer takes orders from other elevators, but takes only its own orders(both internal and external)
		case NoNetwork:
			Println("Current state: noNetwork")
			stateUpdate <- state
			stateUpdate2 := make(chan int)
			peerChannels.PeerStatusUpdate <- PeerStatus{myID, true}
			numberOfPeers := 0
			masterID := -1
			regainedNetworkConnection := false
			doneWithOrders := false
			stateUpdateDelay := time.NewTimer(45 * time.Millisecond)
			stateUpdateDelay.Stop()
			go BybassNetwork(myID, stateUpdate2, sendChannels, recieveChannels)
			for state == NoNetwork {
				select {
				case p := <-peerChannels.PeerUpdateCh:
					if numberOfPeers == 0 {
						regainedNetworkConnection = true
						if doneWithOrders && regainedNetworkConnection {
							stateUpdateDelay.Reset(100 * time.Millisecond)
						}
					}
					numberOfPeers = len(p.Peers)
				case status := <-statusReportsSend3:
					if status.Timeout {
						state = DeadElevator
					}
					if status.Idle && !status.DoorOpen {
						doneWithOrders = true
						if doneWithOrders && regainedNetworkConnection {
							stateUpdateDelay.Reset(100 * time.Millisecond)
						}
					} else {
						if !stateUpdateDelay.Stop() && doneWithOrders && regainedNetworkConnection {
							<-stateUpdateDelay.C
						}
						doneWithOrders = false
					}
				case m := <-peerChannels.MasterBroadcast:
					if len(m.Peers) != 0 {
						masterID, _ = strconv.Atoi(m.Peers[0])
					} else {
						masterID = -1
					}
				case <-stateUpdateDelay.C:
					if masterID == -1 {
						state = Master
						masterIDUpdate <- myID
						Println("New master: ", myID)
						stateUpdate2 <- state
					} else {
						state = Slave
						masterIDUpdate <- masterID
						Println("New master: ", masterID)
						stateUpdate2 <- state
						//pushOrdersToMaster <- true
						//<-pushOrdersToMaster
					}
				}
			}
		//When an elevator is in case DeadElevator will give all orders to the previous master. If that was the elevator itself, it will save the orders for future completion.
		//If the master is another elevator, it will assign external orders to an active elevator, and internal orders to this elevator for future completion.
		case DeadElevator:
			Println("Current state: DeadElevator")
			stateUpdate <- state
			masterID := -1
			peerChannels.PeerTxEnable <- false
			peerChannels.MasterBroadcastEnable <- false
			stateUpdateDelay := time.NewTimer(100 * time.Millisecond)
			stateUpdateDelay.Stop()
			doneWithOrders := false
			for state == DeadElevator {
				select {
				case status := <-statusReportsSend3:
					if !status.Timeout && status.Idle && !status.DoorOpen {
						doneWithOrders = true
						stateUpdateDelay.Reset(100 * time.Millisecond)
					} else {
						if !stateUpdateDelay.Stop() && doneWithOrders {
							<-stateUpdateDelay.C
						}
						doneWithOrders = false
					}
				case <-peerChannels.PeerUpdateCh:
				case m := <-peerChannels.MasterBroadcast:
					if len(m.Peers) != 0 {
						masterID, _ = strconv.Atoi(m.Peers[0])
						Println("Master Update: ", masterID)
						//pushOrdersToMaster <- true
						//<-pushOrdersToMaster
					}
				case <-stateUpdateDelay.C:
					state = Slave
				}
			}
			peerChannels.PeerTxEnable <- true
			p := <-peerChannels.PeerUpdateCh
			numberOfPeers := len(p.Peers)
			if numberOfPeers == 1 {
				state = Master
				Println("New master1 in dead: ", masterID)
				masterIDUpdate <- myID
			} else {
				state = Slave
				if masterID != -1 {
					Println("New master2 in dead: ", masterID)
					masterIDUpdate <- masterID
				}
			}
		}
	}
}
