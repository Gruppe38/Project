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
func LocalElevator(movementInstructions chan ElevatorMovement, statusReport chan ElevatorStatus) {

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
			Println("LocalElevator() got new instruction", instruction)
			targetFloor = instruction.TargetFloor
			nextDir = instruction.NextDir
			if instruction.Dir {
				driver.SetDirection(-1)
			} else {
				driver.SetDirection(1)
			}

			//Hvis vi faktisk er i riktig etasje, må vi åpne dører og cleare ordre
			//Hvordan?
			//Åpne dører er enkelt, kopier kode fra under
			//Cleare ordre skjer bare når status endrer seg og inkluderer at dør er åpen
			//dvs vi må force en status endring?
			if targetFloor == checkSensors() {
				Println("LocalElevator() door opened")
				driver.SetMotor(false)
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetDoor(1)
				if nextDir {
					driver.SetDirection(-1)
				} else {
					driver.SetDirection(1)
				}
			} else if !doorOpen {
				driver.SetMotor(true)
				waitingForDoor = false
				Println("LocalElevator() is not waiting for door to close")
			} else {
				Println("LocalElevator() is waiting for door to close, we are at floor", checkSensors(), "target is", targetFloor)
				waitingForDoor = true
			}

		case floor := <-currentFloorChan:
			Println("LocalElevator() got a floor update", floor)
			if targetFloor == floor {
				Println("LocalElevator() door opened")
				driver.SetMotor(false)
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetDoor(1)
				if nextDir {
					driver.SetDirection(-1)
				} else {
					driver.SetDirection(1)
				}
			}
		case <-doorTimer.C:
			Println("LocalElevator() door closed")
			doorOpen = false
			driver.SetDoor(0)
			if waitingForDoor {
				Println("LocalElevator() starting motor after timer is done")
				driver.SetMotor(true) 
				waitingForDoor = false
			}
		}
	}
}

//currentFloorChan, sender til localElevator
//statusReport, sender til boradcastElevatorStatus
func watchElevator(currentFloorChan chan int, statusReport chan ElevatorStatus) {
	last := -1
	quit := false
	timeout := false
	lastDir := false
	doorOpen := false
	var status ElevatorStatus
	watchDog := time.NewTimer(5 * time.Second)
	watchDog.Stop()
	for !quit {
		select {
		case <-watchDog.C:
			timeout = true
			lastDir = driver.GetMotor()==1
			doorOpen = driver.GetDoorStatus()
			status = ElevatorStatus{lastDir, last, !timeout, false, doorOpen}
			statusReport <- status
		default:
			i := checkSensors()
			switch i {
			case last:
				break
			default:
				currentFloorChan <- i
				lastDir = driver.GetMotor()==1
				doorOpen = driver.GetDoorStatus()
				idle := driver.GetIdle()
				if i == -1 {
					watchDog.Reset(5 * time.Second)
					status = ElevatorStatus{lastDir, last, !timeout, idle, doorOpen}
				} else {
					if !watchDog.Stop() && !timeout && i == -1 {
						<-watchDog.C
					}
					timeout = false
					status = ElevatorStatus{lastDir, i, !timeout, idle, doorOpen}
				}
				last = i
				statusReport <- status
			}
			if (lastDir != (driver.GetMotor()==1)) || (doorOpen != driver.GetDoorStatus()) {
				lastDir = driver.GetMotor()==1
				doorOpen = driver.GetDoorStatus()
				idle := driver.GetIdle()
				Println("watchElevator() UPDATING STATUS DUE TO DOOR (",doorOpen,") OR MOTORDIR (",lastDir,")" )
				status = ElevatorStatus{lastDir, i, !timeout, idle, doorOpen}
				statusReport <- status
			}
		}
	}
}

func checkSensors() int {
	/*if driver.ReadBit(SENSOR1) {
		return 1
	}
	if driver.ReadBit(SENSOR2) {
		return 2
	}
	if driver.ReadBit(SENSOR3) {
		return 3
	}
	if driver.ReadBit(SENSOR4) {
		return 4
	}
	return -1*/
	return driver.GetFloorSignal()
}

//Som timer 2, men isteded tar inn bool, og returnerer om status er samme som bool, på samme kanal
//Fordel: Mer "logisk" bruk av channel, brukes som mer enn bare trigger til et event
/*
func timer3(start chan bool, ask chan bool, shutdownChan chan bool){
	doorTimer := time.NewTimer(3*time.Second)
	doorTimer.Stop()
	currentlyRunning := false
	for{
		select{
		case  <- start:
			//Avoid race condition
			if !doorTimer.Stop() && currentlyRunning{
				<- doorTimer.C
			}
			doorTimer.Reset(3*time.Second)
			currentlyRunning = true
		case question := <- ask:
			ask <- question == currentlyRunning
		case  <- doorTimer.C :
			currentlyRunning = false
		case  <- shutdownChan:
			break
		}
	}
}
*/

//buttons sender til watchIncommingOrders
func MonitorOrderbuttons(buttons chan int) {
	last := -1
	for {
		noButtonsPressed := true
		for i := 0; i < N_FLOOR; i++ {
			for j := 0; j < 3; j++ {
				if !(i == 0 && j == 1) && !(i == N_FLOOR-1 && j == 0) {
					currentButton := OrderButtonMatrix[i][j]
					if driver.GetButtonSignal(j,i) {
						noButtonsPressed = false
						if currentButton != last {
							Println("New button: ", currentButton)
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
