package driver

/*
#cgo LDFLAGS: -lcomedi -lm -std=c99
#include "io.h"
*/
import "C"
import . "../../defs/"

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
	success := bool(int(C.sim_init()) == 1)
	return success
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
