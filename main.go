package main

import (
	"driver"
	. "fmt"
)

LIGHT_IND := {FLOOR_IND1, FLOOR_IND2, FLOOR_IND3, FLOOR_IND4}

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

func main() {
	Println("test")

	driver.io_clear_bit(MOTORDIR)
	driver.WriteAnalog(MOTOR, 2800)
	last := 0
	for {
		i := checkSensors()
		switch i {
			case last:
				break
			case -1: 
				driver.clearBit(LIGHT_IND[last-1])
			default:
				driver.SetBit(LIGHT_IND[i-1])
		}
		last = i

		if driver:ReadBit(OBSTRUCTION) {
			driver.WriteAnalog(MOTOR, 0)
		}
		else{	
			if driver.ReadBit(SENSOR1) {
				//Drive up
				driver.io_clear_bit(MOTORDIR)
				driver.WriteAnalog(MOTOR, 2800)
			}

			if driver.ReadBit(SENSOR4) {
				//Drive down
				driver.io_clear_bit(MOTORDIR)
				driver.WriteAnalog(MOTOR, 2800)
			}
		}
	}
}
