package driver

import (
	. "../../defs/"
	"../io/"
	. "fmt"
	"time"
)

//Recieves a movement instruction consisting of direction and target floor, sets motor direction and speed and stops the elevator
//and opens the doors for three seconds when arriving at target floor.
func ExecuteInstructions(movementInstructions <-chan ElevatorMovement, statusReport chan ElevatorStatus, movementReport chan<- ElevatorMovement) {

	currentFloorChan := make(chan int)
	go watchElevator(currentFloorChan, statusReport)

	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	doorIsOpen := false
	waitingForDoor := false
	targetFloor := 0
	nextDir := false

	for {
		select {
		case instruction := <-movementInstructions:
			Println("LocalElevator() got new movementInstruction: ", instruction)
			targetFloor = instruction.TargetFloor
			nextDir = instruction.NextDir
			if instruction.Dir {
				driver.SetBit(MOTORDIR)
			} else {
				driver.ClearBit(MOTORDIR)
			}

			//Hvis vi faktisk er i riktig etasje, må vi åpne dører og cleare ordre
			//Hvordan?
			//Åpne dører er enkelt, kopier kode fra under
			//Cleare ordre skjer bare når status endrer seg og inkluderer at dør er åpen
			//dvs vi må force en status endring?
			if targetFloor == checkSensors() {
				movementReport <- ElevatorMovement{instruction.Dir, nextDir, targetFloor}
				//Println("LocalElevator() door opened")
				driver.WriteAnalog(MOTOR, 0)
				if !doorTimer.Stop() && doorIsOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorIsOpen = true
				driver.SetBit(DOOR_OPEN)
				//Println("We have reached targetFloor, doors have opened")
				if nextDir {
					driver.SetBit(MOTORDIR)
				} else {
					driver.ClearBit(MOTORDIR)
				}
			} else if !doorIsOpen {
				driver.WriteAnalog(MOTOR, 2800)
				waitingForDoor = false
				//Println("We have not reached targetFloor, and we are not waiting for doors to close")
			} else {
				//Println("We are waiting for the doors to close before we can move to next targetFloor\n We are at floor: ", checkSensors(), ". Next targetFloor is: ", targetFloor)
				waitingForDoor = true
			}

		case floor := <-currentFloorChan:
			//Println("LocalElevator() got a floor update: ", floor)
			if targetFloor == floor {
				movementReport <- ElevatorMovement{nextDir, nextDir, targetFloor}
				//Println("LocalElevator() door opened")
				driver.WriteAnalog(MOTOR, 0)
				if !doorTimer.Stop() && doorIsOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorIsOpen = true
				driver.SetBit(DOOR_OPEN)
				//Println("We have reached targetFloor, doors have opened")
				if nextDir {
					driver.SetBit(MOTORDIR)
				} else {
					driver.ClearBit(MOTORDIR)
				}
			}
		case <-doorTimer.C:
			doorIsOpen = false
			driver.ClearBit(DOOR_OPEN)
			//Println("Doors have closed")
			if waitingForDoor {
				Println("LocalElevator() starting motor when we are done waiting for doors to close")
				driver.WriteAnalog(MOTOR, 2800)
				waitingForDoor = false
			}
		}
	}
}

//Reports the current floor whenever the elevator arrives at a floor.
//Generates and sends a report whenever the elevator opens a door, turns on/off the motor, changes direction, elevator motor dies and reaches or leaves a floor.
func watchElevator(currentFloorReport chan<- int, statusReport chan<- ElevatorStatus) {
	lastFloor := -1
	timeout := false
	lastDir := false
	doorOpen := false
	atFloor := false
	idle := true
	watchDog := time.NewTimer(5 * time.Second)
	watchDog.Stop()

	for {
		select {
		case <-watchDog.C:
			Println("Timer ran out, timout activated")
			Println("Timer ran out, timout activated")
			Println("Timer ran out, timout activated")
			Println("Timer ran out, timout activated")
			Println("Timer ran out, timout activated")
			timeout = true
			lastDir = driver.ReadBit(MOTORDIR)
			doorOpen = driver.ReadBit(DOOR_OPEN)
			statusReport <- ElevatorStatus{lastDir, lastFloor, timeout, atFloor, false, doorOpen}
			Println("Status update sent, this eleavtor is not seen as active until this status \n is updated (when it has reached a new floor)")
		default:
			currentFloor := checkSensors()
			switch currentFloor {
			case lastFloor:
				break
			default:
				Println("Detected floor change")
				lastDir = driver.ReadBit(MOTORDIR)
				doorOpen = driver.ReadBit(DOOR_OPEN)
				idle = driver.ReadAnalog(MOTOR) == 0
				if currentFloor == -1 {
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					Println("Resetting timer due to leaving floor")
					watchDog.Reset(5 * time.Second)
					atFloor = false
					Println("Elevator status is about to send. Elevator status")
					statusReport <- ElevatorStatus{lastDir, lastFloor, timeout, atFloor, idle, doorOpen}
					Println("Elevator status is sent. Elevator status")
				} else {
					Println("Stopping timer due to arriving at floor")
					if !watchDog.Stop() && !timeout && currentFloor == -1 {
						<-watchDog.C
					}
					timeout = false
					atFloor = true
					Println("Elevator status is about to send. Elevator status")
					statusReport <- ElevatorStatus{lastDir, currentFloor, timeout, atFloor, idle, doorOpen}
					Println("Elevator status is sent. Elevator status")
					currentFloorReport <- currentFloor
					//Println("Current floor is sent to localElevator()")
					setFloorIndicator(currentFloor)
				}
				lastFloor = currentFloor
			}
			lastDirUpdate := driver.ReadBit(MOTORDIR)
			doorOpenUpdate := driver.ReadBit(DOOR_OPEN)
			idleUdpdate := driver.ReadAnalog(MOTOR) == 0
			if lastDir != lastDirUpdate || doorOpen != doorOpenUpdate || idle != idleUdpdate {
				//Println("Checking if direction or doors have changed since last floor update and updating status")
				//Println("Updating status due to non-floor change, idle changed:", idle != idleUdpdate, "dir changed:", lastDir != lastDirUpdate, "door changed:", doorOpen != doorOpenUpdate)
				if !idleUdpdate && idle {
					println("Resetting timer to 5 sec due to statuschange")
					watchDog.Reset(5 * time.Second)
				} else if idleUdpdate && !idle {
					println("Stopping timer due to statuschange")
					if !watchDog.Stop() && !timeout && currentFloor == -1 {
						<-watchDog.C
					}
					timeout = false
				}
				lastDir = lastDirUpdate
				doorOpen = doorOpenUpdate
				idle = idleUdpdate
				//Println("watchElevator() UPDATING STATUS DUE TO DOOR (", doorOpen, ") OR MOTORDIR (", lastDir, ")")
				statusReport <- ElevatorStatus{lastDir, lastFloor, timeout, atFloor, idle, doorOpen}
			}
		}
	}
}

func checkSensors() int {
	if driver.ReadBit(SENSOR1) {
		return 0
	}
	if driver.ReadBit(SENSOR2) {
		return 1
	}
	if driver.ReadBit(SENSOR3) {
		return 2
	}
	if driver.ReadBit(SENSOR4) {
		return 3
	}
	return -1
	//return driver.GetFloorSignal()
}

//Detects if a button is pushed and passes that button on a channel
func MonitorOrderbuttons(buttons chan<- int) {
	last := -1
	for {
		noButtonsPressed := true
		for i := 0; i < N_FLOORS; i++ {
			for j := 0; j < 3; j++ {
				if !(i == 0 && j == 1) && !(i == N_FLOORS-1 && j == 0) {
					currentButton := OrderButtonMatrix[i][j]
					if driver.ReadBit(currentButton) {
						noButtonsPressed = false
						if currentButton != last {
							Println("New button: ", BtoS(currentButton))
							buttons <- currentButton
							last = currentButton
						}
					}
				}
			}
		}
		if noButtonsPressed {
			last = -1
		}
	}
}

// Binary encoding translating a decimal number into a binnary number
func setFloorIndicator(floor int) {
	if 0 <= floor && floor < N_FLOORS {
		if floor > 1 {
			driver.SetBit(FLOOR_IND1)
		} else {
			driver.ClearBit(FLOOR_IND1)
		}
		if floor == 1 || floor == 3 {
			driver.SetBit(FLOOR_IND2)
		} else {
			driver.ClearBit(FLOOR_IND2)
		}
	}
}

func ToggleLights(confirmedQueue map[int]bool) {
	for button, value := range confirmedQueue {
		i, j := GetButtonIndex(button)
		light := LightMatrix[i][j]
		if value {
			driver.SetBit(light)
		} else {
			driver.ClearBit(light)
		}
	}
}
