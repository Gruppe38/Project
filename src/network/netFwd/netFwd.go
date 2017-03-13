package netFwd

import (
	. "../../defs/"
	"../bcast"
	"log"
	//"math/rand"
	"net"
	. "strconv"
	"strings"
	//"time"
)

//todo
//resend unconfirmed messages

//myID, får id til heisen, må fåes rett etter oppstart, fra main
//masterID - tilsvarende myID

func SendToNetwork(me int, masterID <-chan int, stateUpdate chan int, channels SendChannels) {
	master := <-masterID
	state := <-stateUpdate
	var messageCounter int64 = 0

	//unconfirmedStatus := make(map[int64]StatusMessage)
	//unconfirmedButton := make(map[int64]ButtonMessage)
	//unconfirmedOrders := make(map[int64]OrderMessage)

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessageNet)
	ackRx := make(chan AckMessage)

	go bcast.Transmitter(13038, statusMes, buttonMes, ordersMes)
	go bcast.Receiver(14038, ackRx)
	for {
		switch state {
		case Master, Slave:
			println("Network sender in state slave or master")
			for state == Master || state == Slave {
				select {
				case state = <-stateUpdate:
					break
				case stat := <-channels.Status:
					println("SENDING NEW STATUS OVER NETWORK")
					//println("SendToNetwork() got status dir:", stat.Dir, " lastFloor:", stat.LastFloor, " activeMotor:", stat.ActiveMotor, " atFloor", stat.AtFloor, "doorOpen", stat.DoorOpen)
					messageID := messageCounter
					messageCounter++
					statMes := StatusMessage{stat, me, master, messageID}
					//unconfirmedStatus[messageID] = statMes
					statusMes <- statMes
					//println("SendToNetwork() sent status dir:", stat.Dir," lastFloor: ", stat.LastFloor, " activeMotor:" ,stat.ActiveMotor, " atFloor", stat.AtFloor, "doorOpen", stat.DoorOpen)
				case button := <-channels.ButtonNew:
					//println("SendToNetwork() got new button", button)
					messageID := messageCounter
					messageCounter++
					butMes := ButtonMessage{button, true, me, master, messageID}
					//unconfirmedButton[messageID] = butMes
					buttonMes <- butMes
					//println("SendToNetwork() sent new button", button)
				case button := <-channels.ButtonCompleted:
					//println("SendToNetwork() got completed button", button)
					messageID := messageCounter
					messageCounter++
					butMes := ButtonMessage{button, false, me, master, messageID}
					//unconfirmedButton[messageID] = butMes
					buttonMes <- butMes
					//println("SendToNetwork() sent completed button", button)
				case order := <-channels.Orders:
					println("SendToNetwork() got order")
					orderNet := *NewOrderQueueNet()
					for i := 0; i < 3; i++ {
						for k, v := range order.Elevator[i] {
							orderNet.Elevator[i][Itoa(k)] = v
						}
					}
					messageID := messageCounter
					messageCounter++
					ordMes := OrderMessageNet{orderNet, me, EVERYONE, messageID}
					//unconfirmedOrders[messageID] = ordMes
					ordersMes <- ordMes
					//println("SendToNetwork() sent order with messageID:", ordMes.MessageID)
				case ack := <-ackRx:
					//println("SendToNetwork() recieved ack:", ack.Message, " with type ", ack.Type, " ack for me = ", ack.TargetElevator == me)
					if ack.TargetElevator == me {
						/*switch ack.Type {
						case 0:
							delete(unconfirmedStatus, ack.Message)
						case 1:
							delete(unconfirmedButton, ack.Message)
						case 2:
							delete(unconfirmedOrders, ack.Message)
						}*/
						break
					}
				case master = <-masterID:
					println("SendToNetwork() got new master:", master)
					continue
				}
			}
		case NoNetwork:
			println("Network sender in state NoNetwork")
			for state == NoNetwork {
				state = <-stateUpdate
			}
		}
	}
	//random := rand.Int63()
}

//myID, får id til heisen, må fåes rett etter oppstart, fra main
//Alle kanaler sender til createCUrrentQueue()
func RecieveFromNetwork(me int, stateUpdate chan int, channels RecieveChannels) {
	state := <-stateUpdate

	lastStatusMessageID := [3]int64{}
	//lastOrderMessageID := [3]int64{}

	sentAck := make(map[int64]bool)
	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessageNet)
	ackTx := make(chan AckMessage)

	go bcast.Receiver(13038, statusMes, buttonMes, ordersMes)
	go bcast.Transmitter(14038, ackTx)

	//todo
	//sjekk om melding er til meg
	//Send bekreftelse
	//Videresend på kanal, hvis ikke tidligere mottatt
	for {
		switch state {
		case Slave, Master:
			println("Network reciver in state slave or master")
			for state == Master || state == Slave {
				select {
				case state = <-stateUpdate:
					break
				case stat := <-statusMes:
					println("RecieveFromNetwork() got status from elevator ", stat.ElevatorID, " with target ", stat.TargetElevator)
					if stat.TargetElevator == me || stat.TargetElevator == EVERYONE {
						currentStatusMessageID := stat.MessageID
						ackTx <- AckMessage{currentStatusMessageID, 0, me, stat.ElevatorID}
						if lastStatusMessageID[stat.ElevatorID-1] < currentStatusMessageID {
							lastStatusMessageID[stat.ElevatorID-1] = currentStatusMessageID
							stat.TargetElevator = me
							if !sentAck[currentStatusMessageID] {
								sentAck[currentStatusMessageID] = true
								//println("Trying to send to channel status")
								channels.Status <- stat
								//println("Sent to channel status")
								//println("RecieveFromNetwork() sent status with ID ", stat.MessageID)
							} else {
								//println("RecieveFromNetwork() Already sent ack for status with ID ", stat.MessageID)
							}
						}
					}
				case button := <-buttonMes:
					println("RecieveFromNetwork() got button:", button.Message, " from elevator ", button.ElevatorID, " with target ", button.TargetElevator)
					if button.TargetElevator == me || button.TargetElevator == EVERYONE {
						button.TargetElevator = me
						ackTx <- AckMessage{button.MessageID, 1, me, button.ElevatorID}
						//println("RecieveFromNetwork() sent ack for ", button.MessageID)
						if !sentAck[button.MessageID] {
							sentAck[button.MessageID] = true
							if button.MessageType {
								//println("Trying to send to channel buttonNew")
								channels.ButtonNew <- button
								//println("Sent to channel buttonNew")
								//println("RecieveFromNetwork() forwarded new button", button.Message)
							} else {
								//println("Trying to send to channel buttonCompleted")
								channels.ButtonCompleted <- button
								//println("Sent to channel buttonCompleted")
								//println("RecieveFromNetwork() forwarded completed button", button.Message)
							}
						} else {
							//println("Already sent ack for button with ID", button.MessageID)
						}
					}
				case order := <-ordersMes:
					println("RecieveFromNetwork() got order from elevator ", order.ElevatorID, " with target ", order.TargetElevator)
					if order.TargetElevator == me || order.TargetElevator == EVERYONE {
						order.TargetElevator = me
						ackTx <- AckMessage{order.MessageID, 2, me, order.ElevatorID}
						if !sentAck[order.MessageID] {
							orderNet := *NewOrderQueue()
							for i := 0; i < 3; i++ {
								for k, v := range order.Message.Elevator[i] {
									l, _ := Atoi(k)
									orderNet.Elevator[i][l] = v
								}
							}
							ordersNet := OrderMessage{orderNet, order.ElevatorID, order.TargetElevator, order.MessageID}
							sentAck[order.MessageID] = true
							//println("Trying to send to channel orders")
							channels.Orders <- ordersNet
							//println("Sent to channel orders")
							//println("RecieveFromNetwork() forwarded order with ID ", order.MessageID)
						} else {
							//println("RecieveFromNetwork() Already sent ack for order with ID", order.MessageID)
						}
					}
				}
			}
		case NoNetwork:
			println("Network reciver in state NoNetwork")
			for state == NoNetwork {
				state = <-stateUpdate
			}
		}
	}
	//random := rand.Int63()
}

func DirectTransfer(me int, stateUpdate chan int, send SendChannels, recieve RecieveChannels) {
	println("Direct Transfer started")
	state := NoNetwork
	for state == NoNetwork {
		select {
		case state = <-stateUpdate:
			break
		case status := <-send.Status:
			println("Direct transfer got new status")
			statMes := StatusMessage{status, me, me, 0}
			recieve.Status <- statMes
			println("Direct transfer sent new status")
		case order := <-send.ButtonNew:
			println("Direct transfer got new order")
			statMes := ButtonMessage{order, true, me, me, 0}
			recieve.ButtonNew <- statMes
			println("Direct transfer sent new order")
		case order := <-send.ButtonCompleted:
			println("Direct transfer got completed order")
			statMes := ButtonMessage{order, false, me, me, 0}
			recieve.ButtonCompleted <- statMes
			println("Direct transfer sent completed order")
		case orderQueue := <-send.Orders:
			println("Direct transfer got new orderqueue")
			orderMes := OrderMessage{orderQueue, me, me, 0}
			recieve.Orders <- orderMes
			println("Direct transfer sent new orderqueue")
		}
	}
	println("Direct Transfer ending")
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
