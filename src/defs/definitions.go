package definitions

type HelloMsg struct {
	Message string
	Iter    int
}

type ElevatorMovement struct {
	Dir         bool
	NextDir 	bool
	TargetFloor int
}

type StatusMessage struct {
	Message        ElevatorStatus
	ElevatorID     int
	TargetElevator int
	MessageID      int //set by network module
}

type ElevatorStatus struct {
	Dir         bool
	LastFloor   int
	ActiveMotor bool //kanskje endre navn
	AtFloor     bool
	DoorOpen 	bool
}

/* ElevatorQueue struct {
	Orders [N_FLOOR][3]bool
}

type OrderQueue struct {
	Elevator1 ElevatorQueue
	Elevator2 ElevatorQueue
	Elevator3 ElevatorQueue
}*/

type ButtonMessage struct {
	Message        int
	ElevatorID     int
	TargetElevator int
	MessageID      int //set by network module
}

type OrderQueue struct {
	Elevator [3]map[int]bool
}

type OrderMessage struct {
	Message        OrderQueue
	ElevatorID     int
	TargetElevator int
	MessageID      int //set by network module
}

const (
	Init         int = 0
	Master       int = 1
	Slave        int = 2
	NoNetwork    int = 3
	DeadElevator int = 4
)

//in port 4
const PORT4 = 3
const OBSTRUCTION = (0x300 + 23)
const STOP = (0x300 + 22)
const BUTTON_COMMAND1 = (0x300 + 21)
const BUTTON_COMMAND2 = (0x300 + 20)
const BUTTON_COMMAND3 = (0x300 + 19)
const BUTTON_COMMAND4 = (0x300 + 18)
const BUTTON_UP1 = (0x300 + 17)
const BUTTON_UP2 = (0x300 + 16)

//in port 1
const PORT1 = 2
const BUTTON_DOWN2 = (0x200 + 0)
const BUTTON_UP3 = (0x200 + 1)
const BUTTON_DOWN3 = (0x200 + 2)
const BUTTON_DOWN4 = (0x200 + 3)
const SENSOR1 = (0x200 + 4)
const SENSOR2 = (0x200 + 5)
const SENSOR3 = (0x200 + 6)
const SENSOR4 = (0x200 + 7)

//out port 3
const PORT3 = 3
const MOTORDIR = (0x300 + 15) //FALSE == OPP, TRUE == NED
const LIGHT_STOP = (0x300 + 14)
const LIGHT_COMMAND1 = (0x300 + 13)
const LIGHT_COMMAND2 = (0x300 + 12)
const LIGHT_COMMAND3 = (0x300 + 11)
const LIGHT_COMMAND4 = (0x300 + 10)
const LIGHT_UP1 = (0x300 + 9)
const LIGHT_UP2 = (0x300 + 8)

//out port 2
const PORT2 = 3
const LIGHT_DOWN2 = (0x300 + 7)
const LIGHT_UP3 = (0x300 + 6)
const LIGHT_DOWN3 = (0x300 + 5)
const LIGHT_DOWN4 = (0x300 + 4)
const DOOR_OPEN = (0x300 + 3)
const FLOOR_IND2 = (0x300 + 1)
const FLOOR_IND1 = (0x300 + 0)

//out port 0
const PORT0 = 1
const MOTOR = (0x100 + 0)

//non-existing ports = (for alignment)
const BUTTON_DOWN1 = -1
const BUTTON_UP4 = -1
const LIGHT_DOWN1 = -1
const LIGHT_UP4 = -1

const N_FLOOR = 4

var LightMatrix = [N_FLOOR][3]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var OrderButtonMatrix = [N_FLOOR][3]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}
