package main

import (
	. "./src/driver/"
	. "fmt"
)

var LIGHT_IND = [4]int{LIGHT_COMMAND1, LIGHT_COMMAND2, LIGHT_COMMAND3, LIGHT_COMMAND4}

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

func main() {
	Println("test")
	IoInit()

	ClearBit(MOTORDIR)
	WriteAnalog(MOTOR, 2800)
	last := 1
	for {
		i := checkSensors()
		switch i {
		case last:
			break
		case -1:
			break
		default:
			ClearBit(LIGHT_IND[last-1])
			SetBit(LIGHT_IND[i-1])
			last = i
		}

		if ReadBit(OBSTRUCTION) {
			WriteAnalog(MOTOR, 0)
		} else {
			WriteAnalog(MOTOR, 2800)
			if ReadBit(SENSOR1) {
				//Drive up
				ClearBit(MOTORDIR)
				WriteAnalog(MOTOR, 2800)
			}

			if ReadBit(SENSOR4) {
				//Drive down
				SetBit(MOTORDIR)
				WriteAnalog(MOTOR, 2800)
			}
		}
	}
}
