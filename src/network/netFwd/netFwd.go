package netFwd

import (
	. "../../defs/"
	"../bcast"
	"math/rand"
	"time"
)

//todo
//resend unconfirmed messages

//myID, får id til heisen, må fåes rett etter oppstart, fra main
//masterID - tilsvarende myID
func SendToNetwork(myID chan int, masterID chan int, status chan ElevatorStatus, buttonNew chan int, buttonCompleted chan int, orders chan OrderQueue) {
	rand.Seed(time.Now().UTC().UnixNano())
	me := <-myID
	master := <-masterID

	//unconfirmedStatus := make(map[int64]StatusMessage)
	//unconfirmedButton := make(map[int64]ButtonMessage)
	//unconfirmedOrders := make(map[int64]OrderMessage)

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessage)
	ackRx := make(chan AckMessage)

	go bcast.Transmitter(13038, statusMes, buttonMes, ordersMes)
	go bcast.Receiver(14038, ackRx)
	println("SendToNetwork startet")
	for {
		select {
		case stat := <-status:
			println("Network got status dir:", stat.Dir, " lastFloor:", stat.LastFloor, " activeMotor:", stat.ActiveMotor, " atFloor", stat.AtFloor, "doorOpen", stat.DoorOpen)
			messageID := rand.Int63()
			statMes := StatusMessage{stat, me, master, messageID}
			//unconfirmedStatus[messageID] = statMes
			statusMes <- statMes
			println("Network sent status dir:", stat.Dir," lastFloor: ", stat.LastFloor, " activeMotor:" ,stat.ActiveMotor, " atFloor", stat.AtFloor, "doorOpen", stat.DoorOpen)
		case button := <-buttonNew:
			println("Network got new button", button)
			messageID := rand.Int63()
			butMes := ButtonMessage{button, true, me, master, messageID}
			//unconfirmedButton[messageID] = butMes
			buttonMes <- butMes
			println("Network sent new button", button)
		case button := <-buttonCompleted:
			println("Network got completed button", button)
			messageID := rand.Int63()
			butMes := ButtonMessage{button, false, me, master, messageID}
			//unconfirmedButton[messageID] = butMes
			buttonMes <- butMes
			println("Network sent completed button", button)
		case order := <-orders:
			println("Network got order", order.Elevator[0], order.Elevator[1], order.Elevator[2])
			messageID := rand.Int63()
			ordMes := OrderMessage{order, me, master, messageID}
			//unconfirmedOrders[messageID] = ordMes
			ordersMes <- ordMes
			println("Network sent order", order.Elevator[0], order.Elevator[1], order.Elevator[2])
		case ack := <-ackRx:
			println("Network recieved ack:", ack.Message, " ack for me = ", ack.TargetElevator == me)
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
		case me = <-myID:
			println("NetowrkSend got new me:", me)
			continue
		case master = <-masterID:
			println("NetowrkSend got new master:", master)
			continue
		}
	}

	//random := rand.Int63()
}

//myID, får id til heisen, må fåes rett etter oppstart, fra main
func RecieveFromNetwork(myID chan int, status chan StatusMessage, buttonNew chan ButtonMessage, buttonCompleted chan ButtonMessage, orders chan OrderMessage) {
	println("RecieveFromNetwork startet")
	sentAck := make(map[int64]bool)

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessage)
	ackTx := make(chan AckMessage)

	go bcast.Receiver(13038, statusMes, buttonMes, ordersMes)
	go bcast.Transmitter(14038, ackTx)

	me := -1
	time.Sleep(45*time.Millisecond)
	ready := false
	for !ready {
		select {
		case me = <-myID:
			ready = true
		case <- statusMes:
			continue
		case <- buttonMes:
			continue
		case <- ordersMes:
			continue
		}
	}
	println("wait time for reciever passed")
	//todo
	//sjekk om melding er til meg
	//Send bekreftelse
	//Videresend på kanal, hvis ikke tidligere mottatt
	for {
		select {
		case stat := <-statusMes:
			println("NetworkRecieve got status dir:", stat.Message.Dir, " lastFloor:", stat.Message.LastFloor, " activeMotor:", stat.Message.ActiveMotor, " atFloor", stat.Message.AtFloor, "doorOpen", stat.Message.DoorOpen, " stat for me = ", stat.TargetElevator == me, " from elevator ", stat.ElevatorID, " with ID ", stat.MessageID)
			if stat.TargetElevator == me {
				ackTx <- AckMessage{stat.MessageID, 0, me, stat.ElevatorID}
				if !sentAck[stat.MessageID] {
					sentAck[stat.MessageID] = true
					status <- stat
					println("NetworkRecieve sent status with ID ", stat.MessageID)
				} else {
					println("Already sent ack for status with ID ", stat.MessageID)
				}
			}
		case button := <-buttonMes:
			println("NetworkRecieve got button:", button.Message, " button for me = ", button.TargetElevator == me, " from elevator ", button.ElevatorID, " with ID ", button.MessageID)
			if button.TargetElevator == me {
				ackTx <- AckMessage{button.MessageID, 1, me, button.ElevatorID}
				if !sentAck[button.MessageID] {
					sentAck[button.MessageID] = true
					if button.MessageType {
						buttonNew <- button
						println("Network forwarded new button", button.Message)
					} else {
						buttonCompleted <- button
						println("Network forwarded completed button", button.Message)
					}
				} else {
					println("Already sent ack for button with ID", button.MessageID)
				}
			}
		case order := <-orders:
			println("NetworkRecieve got order", " order for me = ", order.TargetElevator == me, " from elevator ", order.ElevatorID, " with ID ", order.MessageID)
			if order.TargetElevator == me {
				ackTx <- AckMessage{order.MessageID, 2, me, order.ElevatorID}
				if !sentAck[order.MessageID] {
					sentAck[order.MessageID] = true
					orders <- order
					println("Network forwarded order with ID ", order.MessageID)
				}
			} else{
				println("Already sent ack for order with ID", order.MessageID)
			}
		}
	}

	//random := rand.Int63()
}
