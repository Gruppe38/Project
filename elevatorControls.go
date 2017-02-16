import (
	"time"
)

// Setter motorhastighet og retning, dørlys og timer
// Mangler timerfunksjonalitet

func localElevator (beskjed(dir, targetFloor) chan, shutdown chan){

	timerShutdown := make(chan bool)
	currentFloorShutdown := make(chan bool)
	timerStart := make(chan bool)
	timerStop := make(chan bool)
	currentFloorChan := make(chan int)


	go timer(timerStart, timerStop, timerShutdown)
	go currentFloor(currentFloorChan, currentFloorShutdown)
	targetFloor := 0

	for {
		select {
			case instruction := <- beskjed:
				targetFloor = instruction.targetFloor

				if instruction.dir == 1 { 
					driver.SetBit(MOTORBIT)	
				} else {
					driver.ClearBit(MOTORBIT)	
				}

				if checkSensors() != targetFloor{
					driver.WriteAnalog(MOTOR.2800)
				}

			case floor := <- currentFloorChan:
				if targetFloor == floor {
					driver.WriteAnalog(MOTOR.0)
					doorTimer.Reset(3*time.Second)
				}

			case quit := <- shutdown:
				timerShutdown <- true
				currentFloorShutdown <- true
				break
		}
	}
}



func currentFloor (currentFloorChan chan int, shutdownChan chan bool) {

	last := -1
	for {
		select {
		case shutdown <- shutdownChan:
			break
		default:
			i := checkSensors()
			switch i {
				case last: 
					break
				default:
					currentFloorChan <- i
					last = i

			}
		}
	}
}

func checkSensors() int {
	if ReadBit(SENSOR1) {
		return 1
	}
	if ReadBit(SENSOR2) {
		return 2
	}
	if ReadBit(SENSOR3) {
		return 3
	}
	if ReadBit(SENSOR4) {
		return 4
	}
	return -1
}

func timer(start chan bool, stop chan bool, shutdownChan chan bool){
	doorTimer := time.NewTimer(3*time.Second)
	for{
		select{
		case _ <- start:
			doorTimer.Stop()
			doorTimer.Reset()

		case _ <- doorTimer.C :

		}
	}
	// startes
	// sjekke om den er startet 
}

func checkOrderbuttons(buttons chan int, shutdown chan bool){
	last := -1
	for {
		noButtonsPressed := true
		for i := 0; i < N_FLOOR; i++{ 
			for j := 0; j < 3; j++{
				if ! (i==0 && j==1) && ! (i==N_FLOOR-1 && j==0){
					currentButton := OrderButtonMatrix[i][j]
					if driver.ReadBit(currentButton){
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















