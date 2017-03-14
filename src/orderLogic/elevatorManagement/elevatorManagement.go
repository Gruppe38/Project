package elevatorManagement

import (
	. "../../defs/"
	//"../../driver/io/"
	. "fmt"
	//"time"
)

//statusReport fra broadcastElevator
//OrderMessage fra createOrderQueue via nettverket
//movementInstructions sender til LocalElevator

//When recieving a satus update or an orders update, calculates instructins on how to serve the next order.
func AssignMovementInstruction(statusReport <-chan ElevatorStatus, orders <-chan OrderMessage, movementInstructions chan<- ElevatorMovement) {
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
	orderButtonMatrix := [N_FLOORS][3]bool{}
	for key, value := range orders {
		if value {
			Println("Knapp: ", BtoS(key), " er aktiv")
			empty = false
			i, j := GetButtonIndex(key)
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

//Decides which order should be served next when given a floor and a direction of travel.
func findNextOrder(status ElevatorStatus, orderButtonMatrix [N_FLOORS][3]bool) ElevatorMovement {
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
			//Println("Siden status.AtFloor er true og status.Dir er false kjøres forløkke fra i= ", status.LastFloor, "til i<N_FLOORS")
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
			//Println("Siden status.AtFloor er false og status.Dir er false kjøres forløkke fra i= ", status.LastFloor+1, "til i<N_FLOORS")
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