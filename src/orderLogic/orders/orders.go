package elevatorManagement

import (
	. "../../defs/"
	//"../../driver/io/"
	."../../driver/elevatorControls/"
	. "fmt"
	"time"
)

//Ovvervåke utførte ordre
//behandle utførte og nye ordre

//statusReport inneholder heistatus, kommer fra broadcastElevatorStatus via nettverk
//CompletedOrders Ikke skrevet andre enden enda
//newOrders kommer fra watchIncommingOrders via nettverk
//orderQueueReport sendes via nettverk, til bland annet createCurrentQueue
func CreateOrderQueue(stateUpdate <-chan int, peerUpdate <-chan PeerStatus, statusReport <-chan StatusMessage, completedOrders <-chan ButtonMessage,
	newOrders <-chan ButtonMessage, orderQueueReport chan<- OrderQueue, orderQueueBackup <-chan OrderMessage) {
	orders := *NewOrderQueue()
	activeElevators := [3]bool{}
	elevatorStatus := [3]ElevatorStatus{}
	state := <-stateUpdate
	for {
		switch state {
		case Master, NoNetwork, DeadElevator:
			for i, active := range activeElevators {
				if !active {
					for order, value := range orders.Elevator[i] {
						if value && order != BUTTON_COMMAND1 && order != BUTTON_COMMAND2 && order != BUTTON_COMMAND3 && order != BUTTON_COMMAND4 {
							println("CreateOrderQueue assigned", BtoS(order), "to: ", i+1)
							cheapestCost := 9999
							cheapestElevator := -1
							for i, v := range activeElevators {
								println("elevator #", i+1, "active =", v)
								if v {
									currentElevatorCost := calculateCost(orders.Elevator[i], elevatorStatus[i], order)
									Println("Cost for elevator", i+1, "is ", currentElevatorCost)
									if currentElevatorCost < cheapestCost {
										cheapestCost = currentElevatorCost
										cheapestElevator = i
									}
								}
							}
							if cheapestElevator == -1 {
								Println(BtoS(order), "not assigned")
								break
							}
							println("CreateOrderQueue assigned", BtoS(order), "to: ", i+1)
							orders.Elevator[cheapestElevator][order] = true
						} else if value {
							orders.Elevator[i][order] = true
						}
					}
				}
			}
			ordersCopy := *NewOrderQueue()
			copy(&orders, &ordersCopy)
			orderQueueReport <- ordersCopy
			for state == Master || state == NoNetwork {
				select {
				case state = <-stateUpdate:
					println("CreateOrderQueue() was told to switch state to", state, "while master")
					break
				case peer := <-peerUpdate:
					Println("new peer update as master", peer)
					activeElevators[peer.ID-1] = peer.Status
				case status := <-statusReport:
					//Println("recieved statusReport in createOrderQueue(): from elevator", status.ElevatorID)
					elevatorStatus[status.ElevatorID-1] = status.Message
				case order := <-completedOrders:
					Println("recieved completedOrders in createOrderQueue(): ", order.Message, " from elevator ", order.ElevatorID)
					i, _ := GetButtonIndex(order.Message)
					orders.Elevator[order.ElevatorID-1][order.Message] = false
					orders.Elevator[order.ElevatorID-1][OrderButtonMatrix[i][2]] = false
					ordersCopy := *NewOrderQueue()
					copy(&orders, &ordersCopy)
					orderQueueReport <- ordersCopy
					Println("Ordre har blitt clearet og oppdatert orderQueue har blitt sendt")
				case order := <-newOrders:
					println("CreateOrderQueue got button(): ", BtoS(order.Message))
					if order.Message == BUTTON_COMMAND1 || order.Message == BUTTON_COMMAND2 ||
						order.Message == BUTTON_COMMAND3 || order.Message == BUTTON_COMMAND4 {
						//println("newOrder is internal button")
						orders.Elevator[order.ElevatorID-1][order.Message] = true
						println("CreateOrderQueue assigned order to: ", order.ElevatorID)
						ordersCopy := *NewOrderQueue()
						copy(&orders, &ordersCopy)
						orderQueueReport <- ordersCopy
					} else {
						cheapestCost := 9999
						cheapestElevator := -1
						//println("newOrder is external button")
						for i, v := range activeElevators {
							println("elevator #", i+1, "active =", v)
							if v {
								currentElevatorCost := calculateCost(orders.Elevator[i], elevatorStatus[i], order.Message)
								Println("Cost for elevator", i+1, "is ", currentElevatorCost)
								if currentElevatorCost < cheapestCost {
									cheapestCost = currentElevatorCost
									cheapestElevator = i
								}
							}
						}
						if cheapestElevator == -1 {
							Println("Order not assigned")
							break
						}
						orders.Elevator[cheapestElevator][order.Message] = true
						ordersCopy := *NewOrderQueue()
						copy(&orders, &ordersCopy)
						orderQueueReport <- ordersCopy
						println("CreateOrderQueue assigned order to: ", cheapestElevator+1)
					}
				case <-orderQueueBackup:
				}
			}
		default:
			select {
			case state = <-stateUpdate:
				println("CreateOrderQueue() was told to switch state to", state, "while not master")
				break
			case peer := <-peerUpdate:
				Println("new peer update while not master", peer)
				activeElevators[peer.ID-1] = peer.Status
			case <-statusReport:
			case <-completedOrders:
				Println("Throwing away completed order")
			case <-newOrders:
				Println("Throwing away new order")
			case updatedOrderQueueMessage := <-orderQueueBackup:
				orders = updatedOrderQueueMessage.Message
			}
		}
	}
}

func copy(original *OrderQueue, clone *OrderQueue) {
	*clone = *original
}

func copyMap(original *map[int]bool, clone *map[int]bool) {
	*clone = *original
}

func calculateCost(orders map[int]bool, status ElevatorStatus, button int) int {
	buttonFloor, _ := GetButtonIndex(button)
	cost := 0
	/*if !status.AtFloor {
		cost++
	} else if status.Idle {
		cost += 1
	}*/
	if status.LastFloor < buttonFloor {
		for floor := status.LastFloor; floor < buttonFloor; floor++ {
			cost++
		}
		if status.Dir {
			cost += 5
		}
	}
	if status.LastFloor > buttonFloor {
		for floor := status.LastFloor; floor > buttonFloor; floor-- {
			cost++
		}
		if !status.Dir {
			cost += 5
		}
	}
	return cost
}

func WatchCompletedOrders(statusReport <-chan ElevatorMovement, buttonReports chan<- int) {
	quit := false
	for !quit {
		status, t := <-statusReport
		//Println("WatchCompletedOrders got a status update")
		if t {
			Println("Siden dør er åpen blir det satt i gang clearing av ordre i etasje: ", status.TargetFloor)
			if status.TargetFloor == N_FLOORS-1 {
				buttonReports <- OrderButtonMatrix[3][1]
			} else if status.TargetFloor == 0 {
				buttonReports <- OrderButtonMatrix[0][0]
			} else if status.NextDir {
				buttonReports <- OrderButtonMatrix[status.TargetFloor][1]
			} else {
				buttonReports <- OrderButtonMatrix[status.TargetFloor][0]
			}
		} else {
			quit = true
		}
	}
}

//buttonReports fra MonitorOrderbuttons
//confirmedQueue fra createCurrentQueue
//forwardOrders sender via nettverket, til createOrderQueue
func WatchIncommingOrders(confirmedQueue <-chan map[int]bool, forwardOrders chan int, pushOrdersToMaster chan bool) {
	nonConfirmedQueue := make(map[int]bool)
	confirmedOrders := make(map[int]bool)
	flushTimer := time.NewTimer(100 * time.Millisecond)
	buttonReports := make(chan int)
	go MonitorOrderbuttons(buttonReports)
	for {
		select {
		case button := <-buttonReports:
			if !nonConfirmedQueue[button] && !confirmedOrders[button] {
				nonConfirmedQueue[button] = true
				//nonConfirmedQueue[button] = true
				forwardOrders <- button
				Println("WatchIncommingOrders() sent button: ", BtoS(button))
			} else {
				Println("WatchIncommingOrders() did not send button: ", BtoS(button))
			}
		case confirmedOrders = <-confirmedQueue:
			for k, v := range confirmedOrders {
				if v {
					nonConfirmedQueue[k] = false
				}
			}
			continue
		case <-flushTimer.C:
			nonConfirmedQueue = make(map[int]bool)
			flushTimer.Reset(100 * time.Millisecond)
		case <-pushOrdersToMaster:
			for order, value := range confirmedOrders {
				if value {
					forwardOrders <- order
				}
			}
			pushOrdersToMaster <- true
			/*			for i := 0; i < N_FLOORS; i++ {
						for j := 0; j < 3; j++{
							button := OrderButtonMatrix[i][j]
							if currentQueue[button] {
								nonConfirmedQueue[button] = false
							}
						}
					}*/
		}
	}
}

//orderQueueReports fra CreateOrderQueue via nettverk
//confirmedQueue sender til watchIncommingOrders
func CreateCurrentQueue(orderQueueReports <-chan OrderMessage, confirmedQueue chan<- map[int]bool) {
	currentQueue := make(map[int]bool)
	for {
		select {
		case orderQueue := <-orderQueueReports:

			for i := 0; i < N_FLOORS; i++ {
				for j := 0; j < 2; j++ {
					button := OrderButtonMatrix[i][j]
					currentQueue[button] = false
					for k := 0; k < 3; k++ {
						button := OrderButtonMatrix[i][j]
						if orderQueue.Message.Elevator[k][button] {
							currentQueue[button] = true
						}
						if k == orderQueue.TargetElevator {
							button := OrderButtonMatrix[i][2]
							currentQueue[button] = orderQueue.Message.Elevator[k-1][button]
						}
					}
				}
			}
			ordersCopy := make(map[int]bool)
			copyMap(&currentQueue, &ordersCopy)
			confirmedQueue <- ordersCopy
			ToggleLights(currentQueue)
		}
	}
}