package elevatorManagement

import (
	. "../../defs/"
	"../../driver/io/"
	"math"
)

//Ovvervåke utførte ordre

//behandle utførte og nye ordre

//set dir og floor

//statusReport inneholder heistatus, kommer fra broadcastElevatorStatus via nettverk
//CompletedOrders Ikke skrevet andre enden enda
//newOrders kommer fra watchIncommingOrders via nettverk
//orderQueueReport sendes via nettverk, til bland annet createCurrentQueue
func createOrderQueue(statusReport chan StatusMessage, completedOrders chan ButtonMessage, 
	 					newOrders chan ButtonMessage, orderQueueReport chan OrderQueue){
	orders := OrderQueue{}
	activeElevators := [3]bool{}
	elevatorStatus = [3]ElevatorStatus{}
	for {
		select {
		case status := <- statusReport:
			elevatorStatus[status.ElevatorID-1] = status.Message
			activeElevators[status.ElevatorID-1] = true
		case order := <- completedOrders:
			orders.Elevator[order.ElevatorID-1][order.Message] = false
			orderQueueReport <- orders
		case order := <- newOrders:
			if order.Message == BUTTON_COMMAND1 || order.Message == BUTTON_COMMAND2 || 
			order.Message == BUTTON_COMMAND3 || order.Message == BUTTON_COMMAND4{
				orders.Elevator[order.ElevatorID-1][order.Message]
			} else {
				cheapestCost := 9999
				cheapestElevator := -1
				for i, v := range activeElevators {
					if v {
						if calculateCost(orders[i],elevatorStatus[i], order.Message) < cheapestCost {
							cheapestElevator = i
						}
					}
				}
				if cheapestElevator == -1{
					break
				}
				orders.Elevator[cheapestElevator][order.Message] = true
				orderQueueReport <- orders
			}
		}
	}
}

func calculateCost(orders map[int]bool{}, status ElevatorStatus, button int) int {
	buttonFloor, buttonType := getButtonIndex(button)
	buttonElevatorSameDir := compareButtonAndElevatorDir(buttonType, status.Dir)
	elevatorFloor := -1
	if status.AtFloor {
		elevatorFloor = status.LastFloor
	} else {
		if status.Dir{
			elevatorFloor = status.LastFloor - 1
		} else {
			elevatorFloor = status.LastFloor + 1
		}
	}
	buttonIndex := -1
	switch buttonType {
	case 0:
		buttonIndex = getDistanceIndex(buttonFloor, false)
	case 1:
		buttonIndex = getDistanceIndex(buttonFloor, true)
	case 2:
		buttonIndex = int(math.Min(float64(getDistanceIndex(buttonFloor, false)),float64(getDistanceIndex(buttonFloor, false))))
	}
	distanceCost := math.Abs(getDistanceIndex(elevatorFloor,status.Dir)-buttonIndex)

	//Distansekostnad, 3 per snuing.
}

func getDistanceIndex(floor int, dir bool) int {
	if dir {
		return 7 - floor
	} else {
		return floor -1
	}
}

func compareButtonAndElevatorDir(buttonType int, elevatorDir bool) bool {
	if buttonType == 2 {
		return true
	}
	if elevatorDir {
		return buttonType == 1
	} else {
		return buttonType == 0
	}
}

//statusReport fra broadcastElevator
//OrderMessage fra createOrderQueue via nettverket
//movementInstructions sender til LocalElevator
func Destination(statusReport chan ElevatorStatus, orders chan OrderMessage, movementInstructions chan ElevatorMovement){
	status := ElevatorStatus{}
	myOrders := make(map[int]bool)
	for {
		select{
		case status = <- statusReport:
			instructions := calculateDestination(status, myOrders)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
			}
		case orders := <- OrderMessage:
			myOrders = orders.Message[orders.TargetElevator]
			instructions := calculateDestination(status, myOrders)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
			}
		}
	}
}

func calculateDestination(status ElevatorStatus, orders map[int]bool) ElevatorMovement{
	empty := true
	orderButtonMatrix := [N_FLOOR][3]bool
	for key, value := range orders{
		if value{
			empty = false
			i,j := getButtonIndex(key)
			orderButtonMatrix[i][j] = true
		}
	}
	if empty {
		return ElevatorMovement{status.Dir, status.Dir, -1}
	}
	instructions := findNextOrder(status ElevatorStatus, orderButtonMatrix)
	if instructions.TargetFloor == -1 {
		status.Dir = !status.Dir
		instructions = findNextOrder(status ElevatorStatus, orderButtonMatrix)
	}
	return instructions
}

func findNextOrder(status ElevatorStatus, orderButtonMatrix [N_FLOOR][3]bool) ElevatorMovement{
	switch status.Dir{
	case true:
		if status.AtFloor{
			i:= status.LastFloor
		} else {
			i:= status.LastFloor - 1
		}
		for i;i>=0;i--{
			if orderButtonMatrix[i][1] || orderButtonMatrix[i][2] {
				return ElevatorMovement{status.Dir, status.Dir, i}
			}
		}
		for i;i<status.LastFloor;i++{
			if orderButtonMatrix[i][0] {
				return ElevatorMovement{status.Dir, !status.Dir, i}
			}
		}

	case false:
		if status.AtFloor{
			i:= status.LastFloor
		} else {
			i:= status.LastFloor + 1
		}
		for i;i<N_FLOOR;i++{
			if orderButtonMatrix[i][0] || orderButtonMatrix[i][2] {
				return ElevatorMovement{status.Dir,status.Dir, i}
			}
		}
		for i;i>=status.LastFloor;i--{
			if orderButtonMatrix[i][0] {
				return ElevatorMovement{status.Dir, !status.Dir, i}
			}
		}
	}
	return ElevatorMovement{status.Dir, status.Dir, -1}
}

//statusReport inneholder heistatus, kommer fra watchElevator, som får kannalen via LocalElevator
//Send 1 til createOrderQueue, via nettverk
//send2 til setTargetFloorDIr, ikke skrevet enda. 
func broadcastElevatorStatus(statusReport, send1, send2 chan ElevatorStatus) {
	quit := false**
	for !quit {
		select {
		case status, t := <-statusReport:
			if t {
				send1 <- status
				send2 <- status
			} else {
				send1.Close()
				send2.Close()
				quit = true
			}
		}
	}
}

//buttonReports fra MonitorOrderbuttons
//confirmedQueue fra createCurrentQueue
//forwardOrders sender via nettverket, til createOrderQueue
func watchIncommingOrders(buttonReports chan int, confirmedQueue chan map[int]bool, forwardOrders chan int) {
	currentQueue := make(map[int]bool)
	//Relevant om master forsvinner
	for {
		select {
		case button := <-buttonReports:
			if !currentQueue[button]{
				currentQueue[button] = true
				nonConfirmedQueue[button] = true
				forwardOrders <- button
			}
		case currentQueue := <-confirmedQueue:
			
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
func createCurrentQueue(orderQueueReports chan OrderMessage, confirmedQueue chan map[int]bool){
	for{
		select{
		case orderQueue := <-orderQueueReports:
			for i := 0; i<3; i++ {
				for j := 0; j < N_FLOOR; j++ {
					for k := 0; k < 2; k++{
						button := OrderButtonMatrix[j][k]
						currentQueue = orderQueue.Message.Elevator[i][button]
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
func toggleLights(confirmedQueue map[int]bool){
	for button, value := range confirmedQueue{
		if value {
			driver.SetBit(button)
		} else {
			driver.ClearBit(button)
		}
	}
}

func getButtonIndex(button int) int, int {
	for i := 0; i < N_FLOOR; i++ {
		for j := 0; j < 3; j++{
			if btn == OrderButtonMatrix[i][j]{
				return i, j
			}
		}
	}
}

func getLightIndex(light int) int, int {
	for i := 0; i < N_FLOOR; i++ {
		for j := 0; j < 3; j++{
			if light == LightMatrix[i][j]{
				return i, j
			}
		}
	}
}