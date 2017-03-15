package netFwd

import (
	. "../../defs/"
	"../bcast"
	"log"
	"net"
	. "strconv"
	"strings"
	"time"
)

func SendToNetwork(me int, masterID <-chan int, peerUpdates chan PeerStatus, stateUpdate chan int, channels SendChannels) {
	master := <-masterID
	state := <-stateUpdate
	var messageCounter int64 = 0

	activeElevators := [MAX_ELEVATORS]bool{}
	var recievedAck [MAX_ELEVATORS]map[int64]int
	for elevator, _ := range recievedAck {
		recievedAck[elevator] = make(map[int64]int)
	}

	unconfirmedStatusMessages := make(map[int64]StatusMessage)
	unconfirmedBUttonMessages := make(map[int64]ButtonMessage)
	unconfirmedOrderMessages := make(map[int64]OrderMessageNet)

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessageNet)
	ackRx := make(chan AckMessage)

	go bcast.Transmitter(13038, statusMes, buttonMes, ordersMes)
	go bcast.Receiver(14038, ackRx)
	var resendTicker time.Ticker
	for {
		switch state {
		case Master, Slave, DeadElevator:
			resendTicker = *time.NewTicker(100 * time.Millisecond)
			for state == Master || state == Slave || state == DeadElevator {
				select {
				case state = <-stateUpdate:
					break
				case stat := <-channels.Status:
					messageID := messageCounter
					messageCounter++
					statMes := StatusMessage{stat, me, master, messageID}
					recievedAck[master-1][messageID] = 0
					unconfirmedStatusMessages[messageID] = statMes
					statusMes <- statMes
				case button := <-channels.ButtonNew:
					messageID := messageCounter
					messageCounter++
					butMes := ButtonMessage{button, true, me, master, messageID}
					recievedAck[master-1][messageID] = 1
					unconfirmedBUttonMessages[messageID] = butMes
					buttonMes <- butMes
				case button := <-channels.ButtonCompleted:
					messageID := messageCounter
					messageCounter++
					butMes := ButtonMessage{button, false, me, master, messageID}
					recievedAck[master-1][messageID] = 1
					unconfirmedBUttonMessages[messageID] = butMes
					buttonMes <- butMes
				case order := <-channels.Orders:
					orderNet := *NewOrderQueueNet()
					for i := 0; i < MAX_ELEVATORS; i++ {
						for k, v := range order.Elevator[i] {
							orderNet.Elevator[i][Itoa(k)] = v
						}
					}
					messageID := messageCounter
					messageCounter++
					ordMes := OrderMessageNet{orderNet, me, EVERYONE, messageID}
					for i, v := range activeElevators {
						if v {
							recievedAck[i][messageID] = 2
						}
					}
					unconfirmedOrderMessages[messageID] = ordMes
					ordersMes <- ordMes
				case ack := <-ackRx:
					if ack.TargetElevator == me {
						delete(recievedAck[ack.ElevatorID-1], ack.Message)
					}
				case master = <-masterID:
				case peer := <-peerUpdates:
					activeElevators[peer.ID-1] = peer.Status
					if !peer.Status {
						recievedAck[peer.ID-1] = make(map[int64]int)
					}
				case <-resendTicker.C:
					for elevator, active := range activeElevators {
						if active {
							for messageID, acktype := range recievedAck[elevator] {
								switch acktype {
								case 0:
									temp := unconfirmedStatusMessages[messageID]
									temp.TargetElevator = elevator + 1
									statusMes <- temp
								case 1:
									temp := unconfirmedBUttonMessages[messageID]
									temp.TargetElevator = elevator + 1
									buttonMes <- temp
								case 2:
									temp := unconfirmedOrderMessages[messageID]
									temp.TargetElevator = elevator + 1
									ordersMes <- temp
								}
							}
						}
					}
				}
			}
		case NoNetwork:
			resendTicker.Stop()
			for state == NoNetwork {
				select {
				case master = <-masterID:
				case state = <-stateUpdate:
				case peer := <-peerUpdates:
					activeElevators[peer.ID-1] = peer.Status
					if !peer.Status {
						recievedAck[peer.ID-1] = make(map[int64]int)
					}
				case <-resendTicker.C:
				}

			}
		}
	}
}

func RecieveFromNetwork(me int, masterID <-chan int, stateUpdate chan int, channels RecieveChannels) {
	master := <-masterID
	state := <-stateUpdate

	var sentAck [MAX_ELEVATORS]map[int64]bool
	for elevator, _ := range sentAck {
		sentAck[elevator] = make(map[int64]bool)
	}

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessageNet)
	ackTx := make(chan AckMessage)

	go bcast.Receiver(13038, statusMes, buttonMes, ordersMes)
	go bcast.Transmitter(14038, ackTx)

	for {
		switch state {
		case Slave, Master, DeadElevator:
			for state == Master || state == Slave || state == DeadElevator {
				select {
				case master = <-masterID:
				case state = <-stateUpdate:
				case stat := <-statusMes:
					if stat.TargetElevator == me || stat.TargetElevator == EVERYONE {
						currentStatusMessageID := stat.MessageID
						ackTx <- AckMessage{currentStatusMessageID, 0, me, stat.ElevatorID}
						stat.TargetElevator = me
						if !sentAck[stat.ElevatorID-1][currentStatusMessageID] {
							sentAck[stat.ElevatorID-1][currentStatusMessageID] = true
							channels.Status <- stat
						}
					}
				case button := <-buttonMes:
					if button.TargetElevator == me || button.TargetElevator == EVERYONE {
						button.TargetElevator = me
						ackTx <- AckMessage{button.MessageID, 1, me, button.ElevatorID}
						if !sentAck[button.ElevatorID-1][button.MessageID] {
							sentAck[button.ElevatorID-1][button.MessageID] = true
							if button.MessageType {
								channels.ButtonNew <- button
							} else {
								channels.ButtonCompleted <- button
							}
						}
					}
				case order := <-ordersMes:
					if order.TargetElevator == me || order.TargetElevator == EVERYONE {
						order.TargetElevator = me
						ackTx <- AckMessage{order.MessageID, 2, me, order.ElevatorID}
						if !sentAck[order.ElevatorID-1][order.MessageID] && order.ElevatorID == master {
							orderNet := *NewOrderQueue()
							for i := 0; i < MAX_ELEVATORS; i++ {
								for k, v := range order.Message.Elevator[i] {
									l, _ := Atoi(k)
									orderNet.Elevator[i][l] = v
								}
							}
							ordersNet := OrderMessage{orderNet, order.ElevatorID, order.TargetElevator, order.MessageID}
							sentAck[order.ElevatorID-1][order.MessageID] = true
							channels.Orders <- ordersNet
						}
					}
				}
			}
		case NoNetwork:
			for state == NoNetwork {
				select {
				case master = <-masterID:
				case state = <-stateUpdate:
				}
			}
		}
	}
}

//Transfer values directly from send channels to revice channels
//Used to operate a solo elevator with no network connction.
func BybassNetwork(me int, stateUpdate chan int, send SendChannels, recieve RecieveChannels) {
	state := NoNetwork
	for state == NoNetwork {
		select {
		case state = <-stateUpdate:
			break
		case status := <-send.Status:
			recieve.Status <- StatusMessage{status, me, me, 0}
		case order := <-send.ButtonNew:
			recieve.ButtonNew <- ButtonMessage{order, true, me, me, 0}
		case order := <-send.ButtonCompleted:
			recieve.ButtonCompleted <- ButtonMessage{order, false, me, me, 0}
		case orderQueue := <-send.Orders:
			recieve.Orders <- OrderMessage{orderQueue, me, me, 0}
		}
	}
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")

	return localAddr[0:idx]
}
