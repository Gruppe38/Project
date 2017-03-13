package elevatorManagement

import (
	. "../../defs/"
	"../../driver/io/"
	. "fmt"
	"time"
)

//Ovvervåke utførte ordre

//behandle utførte og nye ordre

//set dir og floor

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
		case Master, NoNetwork:
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
					i, _ := getButtonIndex(order.Message)
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

func calculateCost(orders map[int]bool, status ElevatorStatus, button int) int {
	buttonFloor, _ := getButtonIndex(button)
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

/*func calculateCost(orders map[int]bool, status ElevatorStatus, button int) int {
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
		//Println("Calculating distanceCost for buttontype: UP/0")
		buttonIndex = getDistanceIndex(buttonFloor, false)
		distanceCost = buttonIndex - elevatorIndex
	case 1:
		//Println("Calculating distanceCost for buttontype: DOWN/1")
		buttonIndex = getDistanceIndex(buttonFloor, true)
		distanceCost = buttonIndex - elevatorIndex
	case 2:
		//Println("Calculating distanceCost for buttontype: INTERNAL/2")
		if elevatorIndex-getDistanceIndex(buttonFloor, false) > elevatorIndex-getDistanceIndex(buttonFloor, true) {
			buttonIndex = getDistanceIndex(buttonFloor, true)
		} else {
			buttonIndex = getDistanceIndex(buttonFloor, false)
		}
		distanceCost = buttonIndex - elevatorIndex
	}
	if distanceCost < 0 {
		distanceCost = (N_FLOOR-1)*2 + distanceCost
		Println("adding turncost, due to distanceCost being negative")
		turnCost += 3
	}

	if buttonIndex >= 3 && (elevatorIndex < 3 || elevatorIndex > buttonIndex) {
		Println("adding turncost, buttonIndex being 3 or above, and elevatorIndex not being between that and 3")
		turnCost += 3
	}
	return distanceCost + turnCost

}

func getDistanceIndex(floor int, dir bool) int {
	if dir {
		return 7 - floor
	} else {
		return floor
	}
}*/

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
func Destination(statusReport <-chan ElevatorStatus, orders <-chan OrderMessage, movementInstructions chan<- ElevatorMovement) {
	status := ElevatorStatus{}
	myOrders := make(map[int]bool)
	for {
		select {
		case status = <-statusReport:
			//Println("Recieved statusReport in Destination()")
			instructions := calculateDestination(status, myOrders)
			//Println("Destination() calculated new instructions: ", instructions)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
				//Println("Instructions sent on channel movementInstructions, only if instructions != -1")
			}
		case order := <-orders:
			//Println("Recieved orders in Destination")
			for k, v := range order.Message.Elevator[0] {
				Println("Knapp: ", BtoS(k), " Verdi: ", v)
			}
			myOrders = order.Message.Elevator[order.TargetElevator-1]
			instructions := calculateDestination(status, myOrders)
			//Println("Destination() reciveded instructions", instructions)
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
			Println("Knapp: ", BtoS(key), " er aktiv")
			empty = false
			i, j := getButtonIndex(key)
			orderButtonMatrix[i][j] = true
		}
	}
	if empty {
		//Println("calculateDestination() returenerer instruction med TargetFloor = -1 fordi det ikke finnes noen aktive ordre")
		return ElevatorMovement{status.Dir, status.Dir, -1}
	}
	instructions := findNextOrder(status, orderButtonMatrix)
	//Println("findNextOrder() returnerer: ", instructions)
	if instructions.TargetFloor == -1 {
		status.Dir = !status.Dir
		instructions = findNextOrder(status, orderButtonMatrix)
		//Println("findNextOrder() kjører for andre gang og returnerer: ", instructions)
	}
	return instructions
}

func findNextOrder(status ElevatorStatus, orderButtonMatrix [N_FLOOR][3]bool) ElevatorMovement {
	//Println("Running function findNextOrder()")
	//Println("status.Dir: ", status.Dir)
	//Println("status.AtFloor: ", status.AtFloor)
	switch status.Dir {
	case true:
		//Println("This case runs if status.Dir is true")
		if status.AtFloor {
			//Println("Siden status.AtFloor og status.Dir er true kjøres forløkke fra i= ", status.LastFloor, "til i=0")
			for i := status.LastFloor; i >= 0; i-- {
				if orderButtonMatrix[i][1] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := 0; i < status.LastFloor; i++ {
				if orderButtonMatrix[i][0] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		} else {
			//Println("Siden status.AtFloor er false og status.Dir er true kjøres forløkke fra i= ", status.LastFloor-1, "til i=0")
			for i := status.LastFloor - 1; i >= 0; i-- {
				if orderButtonMatrix[i][1] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := 0; i < status.LastFloor; i++ {
				if orderButtonMatrix[i][0] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		}

	case false:
		if status.AtFloor {
			//Println("Siden status.AtFloor er true og status.Dir er false kjøres forløkke fra i= ", status.LastFloor, "til i<N_FLOOR")
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
			//Println("Siden status.AtFloor er false og status.Dir er false kjøres forløkke fra i= ", status.LastFloor+1, "til i<N_FLOOR")
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
func BroadcastElevatorStatus(statusReport <-chan ElevatorStatus, send1, send2, send3 chan<- ElevatorStatus) {
	quit := false
	for !quit {
		select {
		case status, t := <-statusReport:
			if t {
				Println("Sending 1")
				send1 <- status
				Println("Sending 2")
				send2 <- status
				Println("Sending 3")
				send3 <- status
				Println("sent all")
			} else {
				close(send1)
				close(send2)
				close(send3)
				quit = true
			}
		}
	}
}

func BroadcastOrderMessage(orderMessage <-chan OrderMessage, send1, send2, send3 chan<- OrderMessage) {
	quit := false
	for !quit {
		select {
		case order, t := <-orderMessage:
			if t {
				send1 <- order
				send2 <- order
				send3 <- order
			} else {
				close(send1)
				close(send2)
				quit = true
			}
		}
	}
}

/*func WatchCompletedOrders(statusReport chan ElevatorStatus, buttonReports chan int) {
	quit := false
	for !quit {
		status, t := <-statusReport
		Println("WatchCompletedOrders got a status update")
		if t {
			if status.DoorOpen {
				Println("Siden dør er åpen blir det satt i gang clearing av ordre i etasje: ", status.LastFloor)
				if status.LastFloor == N_FLOOR-1 {
					buttonReports <- OrderButtonMatrix[3][1]
				} else if status.LastFloor == 0 {
					buttonReports <- OrderButtonMatrix[0][0]
				} else if status.Dir {
					buttonReports <- OrderButtonMatrix[status.LastFloor][1]
				} else {
					buttonReports <- OrderButtonMatrix[status.LastFloor][0]
				}
			}
		} else {
			quit = true
		}
	}
}*/

func WatchCompletedOrders(statusReport <-chan ElevatorMovement, buttonReports chan<- int) {
	quit := false
	for !quit {
		status, t := <-statusReport
		//Println("WatchCompletedOrders got a status update")
		if t {
			Println("Siden dør er åpen blir det satt i gang clearing av ordre i etasje: ", status.TargetFloor)
			if status.TargetFloor == N_FLOOR-1 {
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
func WatchIncommingOrders(buttonReports <-chan int, confirmedQueue <-chan map[int]bool, forwardOrders chan int, pushOrdersToMaster chan bool) {
	nonConfirmedQueue := make(map[int]bool)
	confirmedOrders := make(map[int]bool)
	flushTimer := time.NewTimer(100 * time.Millisecond)
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
func CreateCurrentQueue(orderQueueReports <-chan OrderMessage, confirmedQueue chan<- map[int]bool) {
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
			ordersCopy := make(map[int]bool)
			copyMap(&currentQueue, &ordersCopy)
			confirmedQueue <- ordersCopy
			toggleLights(currentQueue)
		}
	}
}

//Help functions
func toggleLights(confirmedQueue map[int]bool) {
	for button, value := range confirmedQueue {
		i, j := getButtonIndex(button)
		light := LightMatrix[i][j]
		if value {
			//Println("Setting light for button", button)
			driver.SetBit(light)
		} else {
			driver.ClearBit(light)
		}
	}
	/*for button, value := range confirmedQueue {
		i,j := getButtonIndex(button)
		if value {
			driver.SetLamp(j,i,1)
		} else {
			driver.SetLamp(j,i,0)
		}
	}*/
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
