package elevatorManagement

import (
	. "../../defs/"
	"../../driver/io/"
)

//Ovvervåke utførte ordre

//behandle utførte og nye ordre

//set dir og floor

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
				if i == -1{
					break
				}
				orders.Elevator[i][order.Message] = true
				orderQueueReport <- orders
			}
		}
	}
}

func calculateCost(orders map[int]bool{}, status ElevatorStatus, button int) int {
	cost := 9999
}


func broadcastElevatorStatus(statusReport, send1, send2 chan ElevatorStatus) {
	quit := false
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

func watchIncommingOrders(buttonReports chan int, confirmedQueue chan map[int]bool, forwardElevatorQueue chan int) {
	currentQueue := make(map[int]bool)
	//Relevant om master forsvinner
	nonConfirmedQueue := make(map[int]bool)
	for {
		select {
		case button := <-buttonReports:
			if !currentQueue[button]|| nonConfirmedQueue[button]{
				currentQueue[button] = true
				nonConfirmedQueue[button] = true
				forwardElevatorQueue <- button
			}
		case orderQueue := <-confirmedQueue:
			currentQueue = orderQueue
			for i := 0; i < N_FLOOR; i++ {
				for j := 0; j < 3; j++{
					button := OrderButtonMatrix[i][j]
					if currentQueue[button] {
						nonConfirmedQueue[button] = false
					}
				}
			}
		}
	}
}


func createCurrentQueue(orderQueueReports chan OrderQueue, send1 chan map[int]bool), send2 chan map[int]bool)){
	for{
		select{
		case orderQueue := <-orderQueueReports:
			for i := 0; i<3; i++ {
				for j := 0; j < N_FLOOR; j++ {
					for k := 0; k < 2; k++{
						button := OrderButtonMatrix[j][k]
						currentQueue = orderQueue.Elevator[i][button]
					}
					//Huske å gjøre hver heis til nr 1 etter mottat melding over nettverk
					if i == 0 {
						button := OrderButtonMatrix[j][2]
						currentQueue[button] = orderQueue.Elevator[i][button]
					}
				}
			}
			send1 <- currentQueue
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