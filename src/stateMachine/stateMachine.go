package stateMachine

import (
	. "../defs/"
	. "../network/netFwd"
	. "fmt"
	"strconv"
	"time"
)

func RunElevator (state int, myID int, stateUpdate chan int, statusReportsSend3 chan ElevatorStatus, masterIDUpdate chan int, pushOrdersToMaster chan bool, peerChannels PeerChannels, sendChannels SendChannels, recieveChannels RecieveChannels) {
	for {
		switch state {
		case Init:
			continue
		case Master:
			Println("Switched state to master")
			stateUpdate <- state
			peerChannels.MasterBroadcastEnable <- true
			for state == Master {
				select {
				case p := <-peerChannels.PeerUpdateCh:
					Printf("Peer update:\n")
					Printf("  Peers:    %q\n", p.Peers)
					Printf("  New:      %q\n", p.New)
					Printf("  Lost:     %q\n", p.Lost)
					if p.New != "" {
						i, _ := strconv.Atoi(p.New)
						if i > 0 {
							Println("newID as int =", i)
							Println("Gained peer", p.New)
							peerChannels.PeerStatusUpdate <- PeerStatus{i, true}
						}
					}
					for _, ID := range p.Lost {
						Println("Lost peer", ID)
						i, _ := strconv.Atoi(ID)
						if i > 0 {
							peerChannels.PeerStatusUpdate <- PeerStatus{i, false}
							Println("Lost peer", i)
							if i == myID {
								println("Detected lost connection")
								state = NoNetwork
							}
						}
					}
				case m := <-peerChannels.MasterBroadcast:
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
			peerChannels.MasterBroadcastEnable <- false
		case Slave:
			Println("Switched state to slave")
			stateUpdate <- state
			p := PeerUpdate{}
			for state == Slave {
				select {
				case m := <-peerChannels.MasterBroadcast:
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
								case p = <-peerChannels.PeerUpdateCh:
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
				case p = <-peerChannels.PeerUpdateCh:
					if p.New != "" {
						i, _ := strconv.Atoi(p.New)
						if i > 0 {
							Println("newID as int =", i)
							Println("Gained peer", p.New)
							peerChannels.PeerStatusUpdate <- PeerStatus{i, true}
						}
					}
					for _, ID := range p.Lost {
						Println("Lost peer", ID)
						i, _ := strconv.Atoi(ID)
						if i > 0 {
							Println("Lost peer", i)
							peerChannels.PeerStatusUpdate <- PeerStatus{i, false}
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
			peerChannels.PeerStatusUpdate <- PeerStatus{myID, true}
			numberOfPeers := 0
			masterID := -1
			stateUpdateDelay := time.NewTimer(45 * time.Millisecond)
			stateUpdateDelay.Stop()
			go BybassNetwork(myID, stateUpdate2, sendChannels, recieveChannels)
			for state == NoNetwork {
				select {
				case p := <-peerChannels.PeerUpdateCh:
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
				case m := <-peerChannels.MasterBroadcast:
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
					/*if numberOfPeers == 1 {
						state = Master
					} else {
						state = Slave
						pushOrdersToMaster <- true
						<-pushOrdersToMaster
					}*/
				}
			}
			stateUpdate2 <- state
			/*Println("Switching state from ", state)
			establishConnection(peerUpdateCh, peerTxEnable, masterIDUpdate,
				masterBroadcast, masterBroadcastEnable, myID, &state)
			Println("Switching state to ", state)*/

		case DeadElevator:
			Println("Entering state deadElevator")
			//Ingen knapper kan betjenes
			//Late som død for nettverket?
			stateUpdate <- state
			peerChannels.PeerTxEnable <- false
			peerChannels.MasterBroadcastEnable <- false
			for state == DeadElevator {
				//WriteAnalog(MOTOR, 0)
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
					}
				}
			}
			peerChannels.PeerTxEnable <- true
			p := <-peerChannels.PeerUpdateCh
			numberOfPeers := len(p.Peers)
			println("Deciding if we are master or not after reverting from timout:", numberOfPeers)
			if numberOfPeers == 1 {
				state = Master
			} else {
				state = Slave
			}
		}
	}
}	