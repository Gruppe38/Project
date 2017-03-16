package elevatorManagement

import (
	. "../../defs/"
)

//When recieving a satus update or an orders update, calculates instructins on how to serve the next order.
func AssignMovementInstruction(statusReport <-chan ElevatorStatus, orders <-chan OrderMessage, movementInstructions chan<- ElevatorMovement) {
	status := ElevatorStatus{}
	status.AtFloor = true
	myOrders := make(map[int]bool)
	for {
		select {
		case status = <-statusReport:
			instructions := calculateDestination(status, myOrders)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
			}
		case order := <-orders:
			myOrders = order.Message.Elevator[order.TargetElevator-1]
			instructions := calculateDestination(status, myOrders)
			if instructions.TargetFloor != -1 {
				movementInstructions <- instructions
			}
		}
	}
}

func calculateDestination(status ElevatorStatus, orders map[int]bool) ElevatorMovement {
	empty := true
	orderButtonMatrix := [N_FLOORS][3]bool{}
	for key, value := range orders {
		if value {
			empty = false
			i, j := GetButtonIndex(key)
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

//Decides which order should be served next when given a floor and a direction of travel.
func findNextOrder(status ElevatorStatus, orderButtonMatrix [N_FLOORS][3]bool) ElevatorMovement {
	switch status.Dir {
	case true:
		if status.AtFloor {
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
			for i := status.LastFloor; i < N_FLOORS; i++ {
				if orderButtonMatrix[i][0] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := N_FLOORS - 1; i >= status.LastFloor; i-- {
				if orderButtonMatrix[i][1] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		} else {
			for i := status.LastFloor + 1; i < N_FLOORS; i++ {
				if orderButtonMatrix[i][0] || orderButtonMatrix[i][2] {
					return ElevatorMovement{status.Dir, status.Dir, i}
				}
			}
			for i := N_FLOORS - 1; i >= status.LastFloor; i-- {
				if orderButtonMatrix[i][1] {
					return ElevatorMovement{status.Dir, !status.Dir, i}
				}
			}
		}
	}
	return ElevatorMovement{status.Dir, status.Dir, -1}
}
