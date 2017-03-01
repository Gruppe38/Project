package elevatorManagement

import (
	. "../../defs/"
)

func broadcastElevatorStatus(recieveStatus, send1, send2 chan ElevatorStatus) {
	quit := false
	for !quit {
		select {
		case status, t := <-recieveStatus:
			if t {
				send1 <- status
				send2 <- status
			} else {
				send1.Close()
				send2.Close()
				quit = !t
			}
		}
	}
}

func ______(buttonReports chan int, orderQueueReports chan OrderQueue, forwardElevatorQueue chan ElevatorQueue) {
	currentQueue := OrderQueue{}
	nonConfirmedQueue := OrderQueue{}
	for {
		select {
		case button := <-buttonReports:

		case orderQueue := <-orderQueueReports:
			continue
		}
	}
}

func getButtonIndex(btn int) int, int {
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