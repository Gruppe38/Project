package driver

import (
	. "../../defs/"
	"../io/"
	. "fmt"
	"time"
)

// Setter motorhastighet og retning, dørlys og timer
// Mangler timerfunksjonalitet

func LocalElevator(movementInstructions chan ElevatorMovement, statusReport chan ElevatorStatus, shutdown chan bool) {

	currentFloorShutdown := make(chan bool)
	currentFloorChan := make(chan int)
	go watchElevator(currentFloorChan, statusReport, currentFloorShutdown)

	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	doorOpen := false
	waitingForDoor := false
	targetFloor := 0
	quit := false

	for !quit {
		select {
		case instruction := <-movementInstructions:
			targetFloor = instruction.TargetFloor
			if instruction.Dir {
				driver.SetBit(MOTORDIR)
			} else {
				driver.ClearBit(MOTORDIR)
			}

			if checkSensors() != targetFloor && !doorOpen {
				driver.WriteAnalog(MOTOR, 2800)
				waitingForDoor = false
			} else {
				waitingForDoor = true
			}

		case floor := <-currentFloorChan:
			Println(floor)
			if targetFloor == floor {
				driver.WriteAnalog(MOTOR, 0)
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetBit(DOOR_OPEN)
			}

		case <-shutdown:
			currentFloorShutdown <- true
			quit = true

		case <-doorTimer.C:
			doorOpen = false
			driver.ClearBit(DOOR_OPEN)
			if waitingForDoor {
				driver.WriteAnalog(MOTOR, 2800)
				waitingForDoor = false
			}
		}
	}
}


func watchElevator(currentFloorChan chan int, statusReport chan ElevatorStatus, shutdownChan chan bool) {
	last := -1
	quit := false
	timeout := false
	var status ElevatorStatus
	watchDog := time.NewTimer(5 * time.Second)
	watchDog.Stop()
	for !quit {
		select {
		case <-shutdownChan:
			quit = true
		case <-watchDog.C:
			timeout = true
			status = ElevatorStatus{driver.ReadBit(MOTORDIR), last, !timeout, false}
			statusReport <- status
		default:
			i := checkSensors()
			switch i {
			case last:
				continue
			default:
				currentFloorChan <- i
				idle := driver.ReadAnalog(MOTOR) == 0
				if i == -1 {
					watchDog.Reset(5 * time.Second)
					status = ElevatorStatus{driver.ReadBit(MOTORDIR), last, !timeout, idle}
				} else {
					if !watchDog.Stop() && !timeout && i == -1 {
						<-watchDog.C
					}
					timeout = false
					status = ElevatorStatus{driver.ReadBit(MOTORDIR), i, !timeout, idle}
				}
				last = i
				statusReport <- status
			}
		}
	}
}

func checkSensors() int {
	if driver.ReadBit(SENSOR1) {
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
	return -1
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

func MonitorOrderbuttons(buttons chan int, shutdown chan bool) {
	last := -1
	for {
		noButtonsPressed := true
		for i := 0; i < N_FLOOR; i++ {
			for j := 0; j < 3; j++ {
				if !(i == 0 && j == 1) && !(i == N_FLOOR-1 && j == 0) {
					currentButton := OrderButtonMatrix[i][j]
					if driver.ReadBit(currentButton) {
						Println(currentButton)
						noButtonsPressed = false
						if currentButton != last {
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
