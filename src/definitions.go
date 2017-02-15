package definitons


type elevatorStatus struct {
	direction bool
	lastFloor int
	isAlive   bool
	atFloor   bool
}

type elevatorQueue struct {
	orders [4][3]bool
}

type orderQueue struct {
	elevator1 elevatorQueue
	elevator2 elevatorQueue
	elevator3 elevatorQueue
}

type orderMessage struc {
	targetElevator int
	message orderQueue
	id int = 0
}