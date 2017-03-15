package elevatorManagement

import (
	. "../../defs/"
	. "../../driver/elevatorControls/"
	"fmt"
	"time"
)

func redistributeOrders(activeElevators [MAX_ELEVATORS]bool, elevatorStatus [MAX_ELEVATORS]ElevatorStatus, orders OrderQueue) OrderQueue {
	for elevator, active := range activeElevators {
		if !active {
			for order, value := range orders.Elevator[elevator] {
				if value && order != BUTTON_COMMAND1 && order != BUTTON_COMMAND2 && order != BUTTON_COMMAND3 && order != BUTTON_COMMAND4 {
					cheapestCost := 9999
					cheapestElevator := -1
					for i, v := range activeElevators {
						if v {
							currentElevatorCost := calculateCost(elevatorStatus[i], order)
							if currentElevatorCost < cheapestCost {
								cheapestCost = currentElevatorCost
								cheapestElevator = i
							}
						}
					}
					if cheapestElevator == -1 {
						break
					}
					orders.Elevator[cheapestElevator][order] = true
					orders.Elevator[elevator][order] = false
				} else if value {
					orders.Elevator[elevator][order] = true
				}
			}
		}
	}
	return orders
}

//Creates and maintains the complete orderQueue containing all assigned orders for each elevator while not a slave.
//While the elevator is a slave it mainatains a copy of the instructions from master.
func CreateOrderQueue(stateUpdate <-chan int, peerUpdate <-chan PeerStatus, statusReport <-chan StatusMessage, completedOrders <-chan ButtonMessage,
	newOrders <-chan ButtonMessage, orderQueueReport chan<- OrderQueue, orderQueueBackup <-chan OrderMessage) {
	orders := *NewOrderQueue()
	activeElevators := [MAX_ELEVATORS]bool{}
	elevatorStatus := [MAX_ELEVATORS]ElevatorStatus{}
	state := <-stateUpdate
	for {
		switch state {
		case Master, NoNetwork, DeadElevator:
			orders = redistributeOrders(activeElevators, elevatorStatus, orders)
			ordersCopy := *NewOrderQueue()
			copy(&orders, &ordersCopy)
			orderQueueReport <- ordersCopy
			for state == Master || state == NoNetwork || state == DeadElevator {
				select {
				case state = <-stateUpdate:
					break
				case peer := <-peerUpdate:
					activeElevators[peer.ID-1] = peer.Status
					if !peer.Status {
						orders = redistributeOrders(activeElevators, elevatorStatus, orders)
						ordersCopy := *NewOrderQueue()
						copy(&orders, &ordersCopy)
						orderQueueReport <- ordersCopy
					}
				case status := <-statusReport:
					elevatorStatus[status.ElevatorID-1] = status.Message
				case order := <-completedOrders:
					i, _ := GetButtonIndex(order.Message)
					orders.Elevator[order.ElevatorID-1][order.Message] = false
					orders.Elevator[order.ElevatorID-1][OrderButtonMatrix[i][2]] = false
					ordersCopy := *NewOrderQueue()
					copy(&orders, &ordersCopy)
					orderQueueReport <- ordersCopy
				case order := <-newOrders:
					if order.Message == BUTTON_COMMAND1 || order.Message == BUTTON_COMMAND2 ||
						order.Message == BUTTON_COMMAND3 || order.Message == BUTTON_COMMAND4 {
						//println("newOrder is internal button")
						orders.Elevator[order.ElevatorID-1][order.Message] = true
						ordersCopy := *NewOrderQueue()
						copy(&orders, &ordersCopy)
						orderQueueReport <- ordersCopy
					} else {
						cheapestCost := 9999
						cheapestElevator := -1
						//println("newOrder is external button")
						for i, v := range activeElevators {
							if v {
								currentElevatorCost := calculateCost(elevatorStatus[i], order.Message)
								if currentElevatorCost < cheapestCost {
									cheapestCost = currentElevatorCost
									cheapestElevator = i
								}
							}
						}
						if cheapestElevator == -1 {
							break
						}
						orders.Elevator[cheapestElevator][order.Message] = true
						ordersCopy := *NewOrderQueue()
						copy(&orders, &ordersCopy)
						orderQueueReport <- ordersCopy
					}
				case <-orderQueueBackup:
				}
			}
		default:
			select {
			case state = <-stateUpdate:
				break
			case peer := <-peerUpdate:
				activeElevators[peer.ID-1] = peer.Status
			case <-statusReport:
			case <-completedOrders:
			case <-newOrders:
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

//Returns the cost for assigning an order to an elevator
func calculateCost(status ElevatorStatus, button int) int {
	buttonFloor, _ := GetButtonIndex(button)
	cost := 0
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

//Recieves a floor and a direction and sends a message which tells master that the order is completed.
func WatchCompletedOrders(movementReport <-chan ElevatorMovement, buttonReports chan<- int) {
	for {
		movement := <-movementReport
		if movement.TargetFloor == N_FLOORS-1 {
			buttonReports <- OrderButtonMatrix[3][1]
		} else if movement.TargetFloor == 0 {
			buttonReports <- OrderButtonMatrix[0][0]
		} else if movement.NextDir {
			buttonReports <- OrderButtonMatrix[movement.TargetFloor][1]
		} else {
			buttonReports <- OrderButtonMatrix[movement.TargetFloor][0]
		}
	}
}

//Recieves a button, checks if it is aware of an order for this button, if not, forward the button to master.
func WatchIncommingOrders(confirmedQueue <-chan map[int]bool, forwardOrders chan<- int, pushOrdersToMaster chan bool) {
	nonConfirmedQueue := make(map[int]bool)
	confirmedOrders := make(map[int]bool)
	flushTimer := time.NewTimer(100 * time.Millisecond)
	forwardButtons := make(chan int)
	go MonitorOrderbuttons(forwardButtons)
	for {
		select {
		case button := <-forwardButtons:
			if !nonConfirmedQueue[button] && !confirmedOrders[button] {
				nonConfirmedQueue[button] = true
				forwardOrders <- button
			}
		case confirmedOrders = <-confirmedQueue:
			for button, value := range confirmedOrders {
				if value {
					nonConfirmedQueue[button] = false
				}
			}
			continue
		//Reenable sending of nonconfirmed buttons
		case <-flushTimer.C:
			nonConfirmedQueue = make(map[int]bool)
			flushTimer.Reset(100 * time.Millisecond)
		//Sends all known orders to master.
		//Used when reconnecting to the network
		case <-pushOrdersToMaster:
			fmt.Println("PUSHING ORDER")
			time.Sleep(500 * time.Millisecond)
			for button, value := range confirmedOrders {
				if value {
					forwardOrders <- button
				}
			}
			for button, value := range nonConfirmedQueue {
				if value {
					forwardOrders <- button
				}
			}
			fmt.Println("PUSHED ORDER pre confirm")
			pushOrdersToMaster <- true
			fmt.Println("PUSHED ORDER post confirm")
		}
	}
}

//Creates a queue containing all active external buttons no matter which elevator they are assigned to
//as well as the active internal buttons for this elevator
func CreateCurrentQueue(orderMessages <-chan OrderMessage, confirmedQueueReport chan<- map[int]bool) {
	currentQueue := make(map[int]bool)
	for {
		select {
		case orders := <-orderMessages:
			for floor := 0; floor < N_FLOORS; floor++ {
				for buttonType := 0; buttonType < 2; buttonType++ {
					button := OrderButtonMatrix[floor][buttonType]
					currentQueue[button] = false
					for elevator := 0; elevator < MAX_ELEVATORS; elevator++ {
						button := OrderButtonMatrix[floor][buttonType]
						if orders.Message.Elevator[elevator][button] {
							currentQueue[button] = true
						}
						if elevator == orders.TargetElevator-1 {
							button := OrderButtonMatrix[floor][2]
							currentQueue[button] = orders.Message.Elevator[elevator][button]
						}
					}
				}
			}
			currentQueueCopy := make(map[int]bool)
			copyMap(&currentQueue, &currentQueueCopy)
			confirmedQueueReport <- currentQueueCopy
			ToggleLights(currentQueue)
		}
	}
}
