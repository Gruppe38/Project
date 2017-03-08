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

	unconfirmedStatus := make(map[int64]StatusMessage)
	unconfirmedButton := make(map[int64]ButtonMessage)
	unconfirmedOrders := make(map[int64]OrderMessage)

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessage)
	ackRx := make(chan AckMessage)

	go bcast.Transmitter(13038, statusMes, buttonMes, ordersMes)
	go bcast.Receiver(14038, ackRx)
	print("SendTONetwork startet")
	for {
		select {
		case stat := <-status:
			messageID := rand.Int63()
			statMes := StatusMessage{stat, me, master, messageID}
			unconfirmedStatus[messageID] = statMes
			statusMes <- statMes
		case button := <-buttonNew:
			print("Network got button", button)
			messageID := rand.Int63()
			butMes := ButtonMessage{button, true, me, master, messageID}
			unconfirmedButton[messageID] = butMes
			buttonMes <- butMes
			print("Network sent button", button)
		case button := <-buttonCompleted:
			messageID := rand.Int63()
			butMes := ButtonMessage{button, false, me, master, messageID}
			unconfirmedButton[messageID] = butMes
			buttonMes <- butMes
		case order := <-orders:
			messageID := rand.Int63()
			ordMes := OrderMessage{order, me, master, messageID}
			unconfirmedOrders[messageID] = ordMes
			ordersMes <- ordMes
		case ack := <-ackRx:
			if ack.TargetElevator == me {
				switch ack.Type {
				case 0:
					delete(unconfirmedStatus, ack.Message)
				case 1:
					delete(unconfirmedButton, ack.Message)
				case 2:
					delete(unconfirmedOrders, ack.Message)
				}
			}
		case me = <-myID:
			continue
		case master = <-masterID:
			continue
		}
	}

	//random := rand.Int63()
}

//myID, får id til heisen, må fåes rett etter oppstart, fra main
func RecieveFromNetwork(myID chan int, status chan StatusMessage, buttonNew chan ButtonMessage, buttonCompleted chan ButtonMessage, orders chan OrderMessage) {
	me := <-myID
	sentAck := make(map[int64]bool)

	statusMes := make(chan StatusMessage)
	buttonMes := make(chan ButtonMessage)
	ordersMes := make(chan OrderMessage)
	ackTx := make(chan AckMessage)

	go bcast.Receiver(13038, statusMes, buttonMes, ordersMes)
	go bcast.Transmitter(14038, ackTx)
	//todo
	//sjekk om melding er til meg
	//Send bekreftelse
	//Videresend på kanal, hvis ikke tidligere mottatt
	for {
		select {
		case stat := <-statusMes:
			if stat.TargetElevator == me {
				ackTx <- AckMessage{stat.MessageID, 0, me, stat.ElevatorID}
				if !sentAck[stat.MessageID] {
					sentAck[stat.MessageID] = true
					status <- stat
				}
			}
		case button := <-buttonMes:
			if button.TargetElevator == me {
				ackTx <- AckMessage{button.MessageID, 1, me, button.ElevatorID}
				if !sentAck[button.MessageID] {
					sentAck[button.MessageID] = true
					if button.MessageType {
						buttonNew <- button
					} else {
						buttonCompleted <- button
					}
				}
			}
		case order := <-orders:
			if order.TargetElevator == me {
				ackTx <- AckMessage{order.MessageID, 2, me, order.ElevatorID}
				if !sentAck[order.MessageID] {
					sentAck[order.MessageID] = true
					orders <- order
				}
			}
		}
	}

	//random := rand.Int63()
}
