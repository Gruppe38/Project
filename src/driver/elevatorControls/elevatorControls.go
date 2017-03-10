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
			Println("LocalElevator() got new movementInstruction: ", instruction)
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
				driver.SetMotor(false)
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetDoor(1)
				Println("We have reached targetFloor, doors hace opened")
				if nextDir {
					driver.SetDirection(-1) 
				} else {
					driver.SetDirection(1)
				}
			} else if !doorOpen {
				driver.SetMotor(true)
				waitingForDoor = false
				Println("We have not reached targetFloor, and we are not waiting for doors to close")
			} else {
				Println("We are waiting for the doors to close before we can move to next targetFloor\n We are at floor: ", checkSensors(), ". Next targetFloor is: ", targetFloor)
				waitingForDoor = true
			}

		case floor := <-currentFloorChan:
			Println("LocalElevator() got a floor update: ", floor)
			if targetFloor == floor {
				driver.SetMotor(false)
				if !doorTimer.Stop() && doorOpen {
					<-doorTimer.C
				}
				doorTimer.Reset(3 * time.Second)
				doorOpen = true
				driver.SetDoor(1)
				Println("We have reached targetFloor, doors have opened")
				if nextDir {
					driver.SetDirection(-1)
				} else {
					driver.SetDirection(1)
				}
			}
		case <-doorTimer.C:
			doorOpen = false
			driver.SetDoor(0)
			Println("Doors have closed")
			if waitingForDoor {
				driver.SetMotor(true) 
				Println("LocalElevator() starting motor when we are done waiting for doors to close")
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
	direction := false
	doorOpen := false
	var status ElevatorStatus
	watchDog := time.NewTimer(5 * time.Second)
	watchDog.Stop()
	for !quit {
		select {
		case <-watchDog.C:
			Println("Elevator has used more than five seconds between two floors")
			timeout = true
			direction = driver.GetMotor()==1
			doorOpen = driver.GetDoorStatus()
			status = ElevatorStatus{direction, last, timeout, false, doorOpen}
			statusReport <- status
			Println("Status update sent, this eleavtor is not seen as active until this status \n is updated (when it has reached a new floor)")
		default:
			i := checkSensors() //i kan ta verdiene 0,1,2,3
			switch i {
			case last:
				break
			default: //Skjer hver gang vi har endret etasje
				currentFloorChan <- i
				Println("Current floor is sent to localElevator()")
				direction = driver.GetMotor()==1
				doorOpen = driver.GetDoorStatus()
				idle := driver.GetIdle()
				if i == -1 {
					watchDog.Reset(5 * time.Second)
					Println("Watchdog is reset")
				} else {
					if !watchDog.Stop() && !timeout && i == -1 {
						<-watchDog.C
					}
					timeout = false
				}
				status = ElevatorStatus{direction, i, timeout, idle, doorOpen}
				last = i
				statusReport <- status
				Println("Elevator status is sent. Elevator status: ", status)
			}
			if (direction != (driver.GetMotor()==1)) || (doorOpen != driver.GetDoorStatus()) {
				Println("Checking if direction or doors have changed since last floor update and updating status")
				direction = driver.GetMotor()==1
				doorOpen = driver.GetDoorStatus()
				idle := driver.GetIdle()
				status = ElevatorStatus{direction, i, timeout, idle, doorOpen}
				statusReport <- status
			}
		}
	}
}

func checkSensors() int {
	/*if driver.ReadBit(SENSOR1) {
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
