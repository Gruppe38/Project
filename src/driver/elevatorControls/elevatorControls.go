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

			if targetFloor == checkSensors() {
				movementReport <- ElevatorMovement{instruction.Dir, nextDir, targetFloor}
				driver.WriteAnalog(MOTOR, 0)
				if !doorTimer.Stop() && doorIsOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorIsOpen = true
				driver.SetBit(DOOR_OPEN)
				if nextDir {
					driver.SetBit(MOTORDIR)
				} else {
					driver.ClearBit(MOTORDIR)
				}
			} else if !doorIsOpen {
				driver.WriteAnalog(MOTOR, 2800)
				waitingForDoor = false
			} else {
				waitingForDoor = true
			}

		case floor := <-currentFloorChan:
			if targetFloor == floor {
				movementReport <- ElevatorMovement{nextDir, nextDir, targetFloor}
				driver.WriteAnalog(MOTOR, 0)
				if !doorTimer.Stop() && doorIsOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorIsOpen = true
				driver.SetBit(DOOR_OPEN)
				if nextDir {
					driver.SetBit(MOTORDIR)
				} else {
					driver.ClearBit(MOTORDIR)
				}
			}
		case <-doorTimer.C:
			doorIsOpen = false
			driver.ClearBit(DOOR_OPEN)
			if waitingForDoor {
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
			timeout = true
			lastDir = driver.ReadBit(MOTORDIR)
			doorOpen = driver.ReadBit(DOOR_OPEN)
			statusReport <- ElevatorStatus{lastDir, lastFloor, timeout, atFloor, false, doorOpen}
		default:
			currentFloor := checkSensors()
			switch currentFloor {
			case lastFloor:
				break
			default:
				lastDir = driver.ReadBit(MOTORDIR)
				doorOpen = driver.ReadBit(DOOR_OPEN)
				idle = driver.ReadAnalog(MOTOR) == 0
				if currentFloor == -1 {
					watchDog.Reset(5 * time.Second)
					atFloor = false
					statusReport <- ElevatorStatus{lastDir, lastFloor, timeout, atFloor, idle, doorOpen}
				} else {
					if !watchDog.Stop() && !timeout && currentFloor == -1 {
						<-watchDog.C
					}
					timeout = false
					atFloor = true
					statusReport <- ElevatorStatus{lastDir, currentFloor, timeout, atFloor, idle, doorOpen}
					currentFloorReport <- currentFloor
					setFloorIndicator(currentFloor)
				}
				lastFloor = currentFloor
			}
			lastDirUpdate := driver.ReadBit(MOTORDIR)
			doorOpenUpdate := driver.ReadBit(DOOR_OPEN)
			idleUdpdate := driver.ReadAnalog(MOTOR) == 0
			if lastDir != lastDirUpdate || doorOpen != doorOpenUpdate || idle != idleUdpdate {
				if !idleUdpdate && idle {
					watchDog.Reset(5 * time.Second)
				} else if idleUdpdate && !idle {
					if !watchDog.Stop() && !timeout && currentFloor == -1 {
						<-watchDog.C
					}
					timeout = false
				}
				lastDir = lastDirUpdate
				doorOpen = doorOpenUpdate
				idle = idleUdpdate
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
