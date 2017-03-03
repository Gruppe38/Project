package elevatorManagement

import (
	. "../../defs/"
	"../../driver/io/"
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
	cost := 9999
	
}

//statusReport fra broadcastElevator
//OrderMessage fra createOrderQueue via nettverket
//movementInstructions sender til LocalElevator
func Destination(statusReport chan ElevatorStatus, orders chan OrderMessage, movementInstructions chan ElevatorMovement){
	//Set target floor og dir for the elevator to complete orders.
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