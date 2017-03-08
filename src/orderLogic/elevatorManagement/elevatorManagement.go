package elevatorManagement

import (
	. "../../defs/"
	"../../driver/io/"
	. "fmt"
)

//Ovvervåke utførte ordre

//behandle utførte og nye ordre

//set dir og floor

//statusReport inneholder heistatus, kommer fra broadcastElevatorStatus via nettverk
//CompletedOrders Ikke skrevet andre enden enda
//newOrders kommer fra watchIncommingOrders via nettverk
//orderQueueReport sendes via nettverk, til bland annet createCurrentQueue
func CreateOrderQueue(statusReport chan StatusMessage, completedOrders chan ButtonMessage,
	newOrders chan ButtonMessage, orderQueueReport chan OrderQueue) {
	orders := OrderQueue{}
	activeElevators := [3]bool{}
	elevatorStatus := [3]ElevatorStatus{}
	for {
		select {
		case status := <-statusReport:
			elevatorStatus[status.ElevatorID-1] = status.Message
			activeElevators[status.ElevatorID-1] = true
		case order := <-completedOrders:
			i, _ := getButtonIndex(order.Message)
			orders.Elevator[order.ElevatorID-1][order.Message] = false
			orders.Elevator[order.ElevatorID-1][OrderButtonMatrix[i][2]] = false
			orderQueueReport <- orders
		case order := <-newOrders:
			if order.Message == BUTTON_COMMAND1 || order.Message == BUTTON_COMMAND2 ||
				order.Message == BUTTON_COMMAND3 || order.Message == BUTTON_COMMAND4 {
				orders.Elevator[order.ElevatorID-1][order.Message] = true
			} else {
				cheapestCost := 9999
				cheapestElevator := -1
				for i, v := range activeElevators {
					if v {
						if calculateCost(orders.Elevator[i], elevatorStatus[i], order.Message) < cheapestCost {
							cheapestElevator = i
						}
					}
				}
				if cheapestElevator == -1 {
					break
				}
				orders.Elevator[cheapestElevator][order.Message] = true
				orderQueueReport <- orders
			}
		}
	}
}

func calculateCost(orders map[int]bool, status ElevatorStatus, button int) int {
	buttonFloor, buttonType := getButtonIndex(button)
	elevatorFloor := -1
	if status.AtFloor {
		elevatorFloor = status.LastFloor
	} else {
		if status.Dir {
			elevatorFloor = status.LastFloor - 1
		} else {
			elevatorFloor = status.LastFloor + 1
		}
	}
	elevatorIndex := getDistanceIndex(elevatorFloor, status.Dir)
	buttonIndex := 0
	distanceCost := 0
	turnCost := 0
	switch buttonType {
	case 0:
		buttonIndex = getDistanceIndex(buttonFloor, false)
		distanceCost = elevatorIndex - buttonIndex
	case 1:
		buttonIndex = getDistanceIndex(buttonFloor, true)
		distanceCost = elevatorIndex - buttonIndex
	case 2:
		if elevatorIndex-getDistanceIndex(buttonFloor, false) > elevatorIndex-getDistanceIndex(buttonFloor, true) {
			buttonIndex = getDistanceIndex(buttonFloor, true)
		} else {
			buttonIndex = getDistanceIndex(buttonFloor, false)
		}
		distanceCost = elevatorIndex - buttonIndex
	}
	if distanceCost < 0 {
		distanceCost = (N_FLOOR-1)*2 + distanceCost
		turnCost += 3
	}

	if buttonIndex >= 3 && (elevatorIndex < 3 || elevatorIndex > buttonIndex) {
		turnCost += 3
	}

	return distanceCost + turnCost

}

func getDistanceIndex(floor int, dir bool) int {
	if dir {
		return 7 - floor
	} else {
		return floor - 1
	}
}

/*func compareButtonAndElevatorDir(buttonType int, elevatorDir bool) bool {
	if buttonType == 2 {
		return true
	}
	if elevatorDir {
		return buttonType == 1
	} else {
		return buttonType == 0
	}
}*/

//statusReport fra broadcastElevator
//OrderMessage fra createOrderQueue via nettverket
//movementInstructions sender til LocalElevator
func Destination(statusReport chan ElevatorStatus, orders chan OrderMessage, movementInstructions chan ElevatorMovement) {
	status := ElevatorStatus{}
	myOrders := make(map[int]bool)
	for {
		select {
		case status = <-statusReport:
			instructions := calculateDestination(status, myOrders)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
			}
		case order := <-orders:
			myOrders = order.Message.Elevator[order.TargetElevator]
			instructions := calculateDestination(status, myOrders)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
			}
		}
	}
}

func calculateDestination(status ElevatorStatus, orders map[int]bool) ElevatorMovement {
	empty := true
	orderButtonMatrix := [N_FLOOR][3]bool{}
	for key, value := range orders {
		if value {
			empty = false
			i, j := getButtonIndex(key)
			orderButtonMatrix[i][j] = true
		}
	}
	if empty {
		return ElevatorMovement{status.Dir, status.Dir, -1}
	}
	instructions := findNextOrder(status, orderButtonMatrix)
	if instructions.TargetFloor == -1 {
		status.Dir = !status.Dir
		instructions = findNextOrder(status, orderButtonMatrix)
	}
	return instructions
}

func findNextOrder(status ElevatorStatus, orderButtonMatrix [N_FLOOR][3]bool) ElevatorMovement {
	switch status.Dir {
	case true:
		if status.AtFloor {
			for i := status.LastFloor; i >= 0; i-- {
				if orderButtonMatrix[i][1] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
				for i := 0; i < status.LastFloor; i++ {
					if orderButtonMatrix[i][0] {
						return ElevatorMovement{status.Dir, !status.Dir, i}
					}
				}
			}
		} else {
			for i := status.LastFloor - 1; i >= 0; i-- {
				if orderButtonMatrix[i][1] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := 4; i < status.LastFloor; i++ {
				if orderButtonMatrix[i][0] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		}

	case false:
		if status.AtFloor {
			for i := status.LastFloor; i < N_FLOOR; i++ {
				if orderButtonMatrix[i][0] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := N_FLOOR - 1; i >= status.LastFloor; i-- {
				if orderButtonMatrix[i][0] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		} else {
			for i := status.LastFloor + 1; i < N_FLOOR; i++ {
				if orderButtonMatrix[i][0] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := N_FLOOR - 1; i >= status.LastFloor; i-- {
				if orderButtonMatrix[i][0] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		}
	}
	return ElevatorMovement{status.Dir, status.Dir, -1}
}

//statusReport inneholder heistatus, kommer fra watchElevator, som får kannalen via LocalElevator
//Send 1 til createOrderQueue, via nettverk
//send2 til destination
//send3 til watchCompletedOrders
func BroadcastElevatorStatus(statusReport, send1, send2, send3 chan ElevatorStatus) {
	quit := false
	for !quit {
		select {
		case status, t := <-statusReport:
			if t {
				send1 <- status
				send2 <- status
				send3 <- status
			} else {
				close(send1)
				close(send2)
				close(send3)
				quit = true
			}
		}
	}
}

func BroadcastOrderMEessage(orderMessage, send1, send2 chan OrderMessage) {
	quit := false
	for !quit {
		select {
		case order, t := <-orderMessage:
			if t {
				send1 <- order
				send2 <- order
			} else {
				close(send1)
				close(send2)
				quit = true
			}
		}
	}
}

func WatchCompletedOrders(statusReport chan ElevatorStatus, buttonReports chan int) {
	quit := false
	for !quit {
		select {
		case status, t := <-statusReport:
			if t {
				if status.DoorOpen {
					if status.LastFloor == 4 {
						buttonReports <- OrderButtonMatrix[3][1]
					} else if status.LastFloor == 1 {
						buttonReports <- OrderButtonMatrix[0][0]
					} else if status.Dir {
						buttonReports <- OrderButtonMatrix[status.LastFloor][0]
					} else {
						buttonReports <- OrderButtonMatrix[status.LastFloor][1]
					}
				}
			} else {
				quit = true
			}
		}
	}
}

//buttonReports fra MonitorOrderbuttons
//confirmedQueue fra createCurrentQueue
//forwardOrders sender via nettverket, til createOrderQueue
func WatchIncommingOrders(buttonReports chan int, confirmedQueue chan map[int]bool, forwardOrders chan int) {
	currentQueue := make(map[int]bool)
	for {
		select {
		case button := <-buttonReports:
			Println("Got button: ", button)
			if !currentQueue[button] {
				currentQueue[button] = true
				//nonConfirmedQueue[button] = true
				forwardOrders <- button
				Println("Sent button: ", button)
			}
		case <-confirmedQueue:
			continue
			/*			for i := 0; i < N_FLOOR; i++ {
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
func CreateCurrentQueue(orderQueueReports chan OrderMessage, confirmedQueue chan map[int]bool) {
	currentQueue := make(map[int]bool)
	for {
		select {
		case orderQueue := <-orderQueueReports:
			for i := 0; i < 3; i++ {
				for j := 0; j < N_FLOOR; j++ {
					for k := 0; k < 2; k++ {
						button := OrderButtonMatrix[j][k]
						currentQueue[button] = orderQueue.Message.Elevator[i][button]
					}
					if i == orderQueue.TargetElevator {
						button := OrderButtonMatrix[j][2]
						currentQueue[button] = orderQueue.Message.Elevator[i][button]
					}
				}
			}
			confirmedQueue <- currentQueue
			toggleLights(currentQueue)
		}
	}
}

//Help functions
func toggleLights(confirmedQueue map[int]bool) {
	for button, value := range confirmedQueue {
		if value {
			driver.SetBit(button)
		} else {
			driver.ClearBit(button)
		}
	}
}

func getButtonIndex(button int) (int, int) {
	for i := 0; i < N_FLOOR; i++ {
		for j := 0; j < 3; j++ {
			if button == OrderButtonMatrix[i][j] {
				return i, j
			}
		}
	}
	return -1, -1
}

func getLightIndex(light int) (int, int) {
	for i := 0; i < N_FLOOR; i++ {
		for j := 0; j < 3; j++ {
			if light == LightMatrix[i][j] {
				return i, j
			}
		}
	}
	return -1, -1
}
