package stateMachine

import (
	. "../defs/"
	. "../network/netFwd"
	"fmt"
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
			fmt.Println("AM SLAVE")
			stateUpdate <- state
			p := PeerUpdate{}
			for state == Slave {
				fmt.Println("READING CHANNELS SLAVE")
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
							if newMaster == myID {
								state = Master
							}
							masterIDUpdate <- newMaster
						}
					}
					for _, ID := range m.New {
						masterID := int(ID)
						masterIDUpdate <- masterID
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
			stateUpdate <- state
			stateUpdate2 := make(chan int)
			peerChannels.PeerStatusUpdate <- PeerStatus{myID, true}
			numberOfPeers := 0
			masterID := -1
			stateUpdateDelay := time.NewTimer(45 * time.Millisecond)
			stateUpdateDelay.Stop()
			go BybassNetwork(myID, stateUpdate2, sendChannels, recieveChannels)
			for state == NoNetwork {
				select {
				case p := <-peerChannels.PeerUpdateCh:
					if numberOfPeers == 0 {
						stateUpdateDelay.Reset(100 * time.Millisecond)
					}
					numberOfPeers = len(p.Peers)
				case status := <-statusReportsSend3:
					if status.Timeout {
						state = DeadElevator
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
						stateUpdate2 <- state
					} else {
						state = Slave
						masterIDUpdate <- masterID
						stateUpdate2 <- state
						//pushOrdersToMaster <- true
						//<-pushOrdersToMaster
					}
				}
			}
		//When an elevator is in case DeadElevator will give all orders to the previous master. If that was the elevator itself, it will save the orders for future completion.
		//If the master is another elevator, it will assign external orders to an active elevator, and internal orders to this elevator for future completion.
		case DeadElevator:
			stateUpdate <- state
			peerChannels.PeerTxEnable <- false
			peerChannels.MasterBroadcastEnable <- false
			for state == DeadElevator {
				select {
				case status := <-statusReportsSend3:
					if !status.Timeout {
						state = Slave
					}
				case <-peerChannels.PeerUpdateCh:
				case m := <-peerChannels.MasterBroadcast:
					if len(m.Peers) != 0 {
						masterID, _ := strconv.Atoi(m.Peers[0])
						masterIDUpdate <- masterID
						//pushOrdersToMaster <- true
						//<-pushOrdersToMaster
					}
				}
			}
			peerChannels.PeerTxEnable <- true
			p := <-peerChannels.PeerUpdateCh
			numberOfPeers := len(p.Peers)
			if numberOfPeers == 1 {
				state = Master
			} else {
				state = Slave
			}
		}
	}
}
