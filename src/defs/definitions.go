package definitions

import (
	"../driver/commBits/"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type ElevatorStatus struct {
	Direction bool
	LastFloor int
	IsAlive   bool
	AtFloor   bool
}

type ElevatorQueue struct {
	Orders [driver.N_FLOOR][3]bool
}

type OrderQueue struct {
	Elevator1 ElevatorQueue
	Elevator2 ElevatorQueue
	Elevator3 ElevatorQueue
}

type OrderMessage struct {
	TargetElevator int
	Message        OrderQueue
	MessageID      int //set by network module
}

const (
	Init         int = 0
	Master       int = 1
	Slave        int = 2
	NoNetwork    int = 3
	DeadElevator int = 4
)
