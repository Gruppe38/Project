package driver

/*
#cgo LDFLAGS: -lcomedi -lm -std=c99
#include "io.h"
#include "elev.h"
*/
import "C"
import . "../../defs/"

var direction = 0
var doorOpen = false
var idle = true

func IoInit() bool {
	success := bool(int(C.io_init()) == 1)

	if success {
		for i := 0; i < N_FLOOR; i++ {
			for j := 0; j < 3; j++ {
				//Avoding down for first floor and up for last floor
				if !(i == 0 && j == 1) && !(i == N_FLOOR-1 && j == 0) {
					ClearBit(LightMatrix[i][j])
				}
			}
		}
		WriteAnalog(MOTOR, 0)
		ClearBit(FLOOR_IND1)
		ClearBit(FLOOR_IND2)
		ClearBit(LIGHT_STOP)
		ClearBit(DOOR_OPEN)
	}
	return success

}

func SimInit() bool {
	C.elev_init()
	return true
}

func SetDirection(val int){
	direction = val
}

func SetMotor(value bool){
	if value {
		C.elev_set_motor_direction(C.int(direction))
	} else {
		C.elev_set_motor_direction(C.int(0))
	}
	idle = !value
}

func GetMotor() int{
	return direction
}

func SetFloorInd(floor int){
	C.elev_set_floor_indicator(C.int(floor))
}

func SetLamp(button int, floor int, value int){
	C.elev_set_button_lamp(C.int(button),C.int(floor),C.int(value))
}

func SetDoor(value int){
	doorOpen = value == 1
	C.elev_set_door_open_lamp(C.int(value))

	println("Setting door bit to",value, "doorOpen is", doorOpen)
}

func GetButtonSignal(button int, floor int) bool {
	return int(C.elev_get_button_signal(C.int(button),C.int(floor))) != 0
}

func GetFloorSignal() int {
	return int(C.elev_get_floor_sensor_signal())
}

func GetDoorStatus() bool {
	return doorOpen
}

func GetIdle() bool {
	return idle
}



func SetBit(channel int) {
	C.io_set_bit(C.int(channel))
}

func ClearBit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func WriteAnalog(channel, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}

func ReadBit(channel int) bool {
	return int(C.io_read_bit(C.int(channel))) != 0
}

func ReadAnalog(channel int) int {
	return int(C.io_read_analog(C.int(channel)))
}
