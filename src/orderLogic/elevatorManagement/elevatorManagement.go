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
	orders := *NewOrderQueue()
	activeElevators := [3]bool{}
	elevatorStatus := [3]ElevatorStatus{}
	Println("Running function CreateOrderQueue")
	for {
		select {
		case status := <-statusReport:
			Println("recieved statusReport in createOrderQueue(): ", status.ElevatorID)
			elevatorStatus[status.ElevatorID-1] = status.Message
			activeElevators[status.ElevatorID-1] = true
		case order := <-completedOrders:
			Println("recieved completedOrders in createOrderQueue(): ", order.Message, " from elevator ", order.ElevatorID)
			i, _ := getButtonIndex(order.Message)
			println("here")
			orders.Elevator[order.ElevatorID-1][order.Message] = false
			println("here")
			orders.Elevator[order.ElevatorID-1][OrderButtonMatrix[i][2]] = false
			println("here")
			orderQueueReport <- orders
		case order := <-newOrders:
			println("CreateOrderQueue got button(): ", order.Message)
			if order.Message == BUTTON_COMMAND1 || order.Message == BUTTON_COMMAND2 ||
						order.Message == BUTTON_COMMAND3 || order.Message == BUTTON_COMMAND4 {
				println("newOrder is internal button")
				orders.Elevator[order.ElevatorID-1][order.Message] = true
				println("CreateOrderQueue assigned order to: ", order.ElevatorID)
				orderQueueReport <- orders
			} else {
				cheapestCost := 9999
				cheapestElevator := -1
				println("newOrder is external button")
				for i, v := range activeElevators {
					println("elevator #",i+1, "active =", v)
					if v {
						if calculateCost(orders.Elevator[i], elevatorStatus[i], order.Message) < cheapestCost {
							cheapestElevator = i
						}
					}
				}
				if cheapestElevator == -1 {
					Println("Order not assigned")
					break
				}
				orders.Elevator[cheapestElevator][order.Message] = true
				orderQueueReport <- orders
				println("CreateOrderQueue assigned order to: ", cheapestElevator+1)
			}
		}
	}
}

func calculateCost(orders map[int]bool, status ElevatorStatus, button int) int {
	buttonFloor, buttonType := getButtonIndex(button)
	elevatorFloor := -1
	Println("Running function calculateCost")
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
		Println("Calculating distanceCost for buttontype: UP/0")
		buttonIndex = getDistanceIndex(buttonFloor, false)
		distanceCost = elevatorIndex - buttonIndex
	case 1:
		Println("Calculating distanceCost for buttontype: DOWN/1")		
		buttonIndex = getDistanceIndex(buttonFloor, true)
		distanceCost = elevatorIndex - buttonIndex
	case 2:
		Println("Calculating distanceCost for buttontype: INTERNAL/2")
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
	Println("Running function getDistanceIndex")
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
	Println("Running function Destination")
	status := ElevatorStatus{}
	myOrders := make(map[int]bool)
	for {
		select {
		case status = <-statusReport:
			//Println("Recieved statusReport in Destination()")
			instructions := calculateDestination(status, myOrders)
			Println("Destination() calculated new instructions: ", instructions)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
				Println("Instructions sent on channel movementInstructions, only if instructions != -1")
			}
		case order := <-orders:
			Println("Recieved orders in Destination")
			for k,v := range(order.Message.Elevator[0]) {
				print(k,v)
			}
			print("\n")
			myOrders = order.Message.Elevator[order.TargetElevator-1]
			instructions := calculateDestination(status, myOrders)
			Println("Destination() reciveded instructions", instructions)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
				Println("Instructions sent on channel movementInstructions, only if instructions != -1")
			}
		}
	}
}

func calculateDestination(status ElevatorStatus, orders map[int]bool) ElevatorMovement {
	//Println("Running function calculateDestination() with orders")
	for k,v := range(orders) {
		print(k,v)
	}
	print("\n")
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
	Println("Running function findNextOrder()")
	Println("status.Dir: ", status.Dir)
	Println("status.AtFloor: ", status.AtFloor)
	Println("status.AtFloor: ", status.LastFloor)
	switch status.Dir {
	case true:
		//Println("This case runs if status.Dir is true")
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
			for i := N_FLOOR-1; i < status.LastFloor; i++ {
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
				if orderButtonMatrix[i][1] {
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
				if orderButtonMatrix[i][1] {
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

func BroadcastOrderMessage(orderMessage, send1, send2 chan OrderMessage) {
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
		status, t := <-statusReport
		Println("WatchCompletedOrders got a status update")
		if t {
			if status.DoorOpen {
				Println("WatchCompletedOrders() clearing orders at floor", status.LastFloor)
				if status.LastFloor == N_FLOOR-1 {
					buttonReports <- OrderButtonMatrix[3][1]
				} else if status.LastFloor == 0 {
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

//buttonReports fra MonitorOrderbuttons
//confirmedQueue fra createCurrentQueue
//forwardOrders sender via nettverket, til createOrderQueue
func WatchIncommingOrders(buttonReports chan int, confirmedQueue chan map[int]bool, forwardOrders chan int) {
	currentQueue := make(map[int]bool)
	knownOrders := make(map[int]bool)
	for {
		select {
		case button := <-buttonReports:
			if !currentQueue[button] && !knownOrders[button] {
				currentQueue[button] = true
				//nonConfirmedQueue[button] = true
				forwardOrders <- button
				Println("WatchIncommingOrders() sent button: ", button)
			} else {
				Println("WatchIncommingOrders() did not send button: ", button)
			}
		case knownOrders = <-confirmedQueue:
			for k,v:= range(knownOrders){
				if v {
					currentQueue[k] = false
				}
			}
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

			for i := 0; i < N_FLOOR; i++ {
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
