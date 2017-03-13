package driver

import (
	. "../../defs/"
	"../io/"
	. "fmt"
	"time"
)

// Setter motorhastighet og retning, dørlys og timer
//movementInstructions fra Destination
//statusReport brukes ikke av denne funkjsonen
func LocalElevator(movementInstructions <-chan ElevatorMovement, statusReport chan ElevatorStatus, movementReport chan<- ElevatorMovement) {

	currentFloorChan := make(chan int)
	go watchElevator(currentFloorChan, statusReport)

	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	doorOpen := false
	waitingForDoor := false
	targetFloor := 0
	nextDir := false
	quit := false

	for !quit {
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
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetBit(DOOR_OPEN)
				//Println("We have reached targetFloor, doors have opened")
				if nextDir {
					driver.SetBit(MOTORDIR)
				} else {
					driver.ClearBit(MOTORDIR)
				}
			} else if !doorOpen {
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
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetBit(DOOR_OPEN)
				//Println("We have reached targetFloor, doors have opened")
				if nextDir {
					driver.SetBit(MOTORDIR)
				} else {
					driver.ClearBit(MOTORDIR)
				}
			}
		case <-doorTimer.C:
			doorOpen = false
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

//currentFloorChan, sender til localElevator
//statusReport, sender til boradcastElevatorStatus
func watchElevator(currentFloorChan chan<- int, statusReport chan<- ElevatorStatus) {
	last := -1
	quit := false
	timeout := false
	lastDir := false
	doorOpen := false
	atFloor := false
	idle := true
	var status ElevatorStatus
	watchDog := time.NewTimer(5 * time.Second)
	watchDog.Stop()
	for !quit {
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
			status = ElevatorStatus{lastDir, last, timeout, atFloor, false, doorOpen}
			statusReport <- status
			Println("Status update sent, this eleavtor is not seen as active until this status \n is updated (when it has reached a new floor)")
		default:
			i := checkSensors()
			switch i {
			case last:
				break
			default:
				Println("Detected floor change")
				lastDir = driver.ReadBit(MOTORDIR)
				doorOpen = driver.ReadBit(DOOR_OPEN)
				idle = driver.ReadAnalog(MOTOR) == 0
				if i == -1 {
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
					status = ElevatorStatus{lastDir, last, timeout, atFloor, idle, doorOpen}
				} else {
					Println("Stopping timer due to arriving at floor")
					if !watchDog.Stop() && !timeout && i == -1 {
						<-watchDog.C
					}
					timeout = false
					atFloor = true
					status = ElevatorStatus{lastDir, i, timeout, atFloor, idle, doorOpen}
					currentFloorChan <- i
					//Println("Current floor is sent to localElevator()")
					setFloorIndicator(i)
				}
				last = i
				statusReport <- status
				//Println("Elevator status is sent. Elevator status: ", status)
			}
			lastDirUpdate := driver.ReadBit(MOTORDIR)
			doorOpenUpdate := driver.ReadBit(DOOR_OPEN)
			idleUdpdate := driver.ReadAnalog(MOTOR) == 0
			if lastDir != lastDirUpdate || doorOpen != doorOpenUpdate || idle != idleUdpdate {
				//Println("Checking if direction or doors have changed since last floor update and updating status")
				//Println("Updating status due to non-floor change, idle changed:", idle != idleUdpdate, "dir changed:", lastDir != lastDirUpdate, "door changed:", doorOpen != doorOpenUpdate)
				lastDir = lastDirUpdate
				doorOpen = doorOpenUpdate
				idle = idleUdpdate
				/*if !idle {
					println("Resetting timer to 5 sec due to statechange")
					watchDog.Reset(5 * time.Second)
				} else {
					println("Stopping timer due to statechange")
					if !watchDog.Stop() && timeout && i == -1 {
						<-watchDog.C
					}
					timeout = false
				}*/
				//Println("watchElevator() UPDATING STATUS DUE TO DOOR (", doorOpen, ") OR MOTORDIR (", lastDir, ")")
				status = ElevatorStatus{lastDir, i, timeout, atFloor, idle, doorOpen}
				statusReport <- status
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

//buttons sender til watchIncommingOrders
func MonitorOrderbuttons(buttons chan<- int) {
	last := -1
	for {
		noButtonsPressed := true
		for i := 0; i < N_FLOOR; i++ {
			for j := 0; j < 3; j++ {
				if !(i == 0 && j == 1) && !(i == N_FLOOR-1 && j == 0) {
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

func setFloorIndicator(floor int) {
	// Binary encoding. One light must always be on.
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
