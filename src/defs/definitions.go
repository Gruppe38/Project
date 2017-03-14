package definitions

type PeerChannels struct {
	PeerUpdateCh chan PeerUpdate
	PeerTxEnable chan bool
	PeerStatusUpdate chan PeerStatus
	MasterBroadcast chan PeerUpdate
	MasterBroadcastEnable chan bool
}

type PeerStatus struct {
	ID     int
	Status bool //true= new, false = lost
}

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

type SendChannels struct {
	Status          chan ElevatorStatus
	ButtonNew       chan int
	ButtonCompleted chan int
	Orders          chan OrderQueue
}

type RecieveChannels struct {
	Status          chan StatusMessage
	ButtonNew       chan ButtonMessage
	ButtonCompleted chan ButtonMessage
	Orders          chan OrderMessage
}

type AckMessage struct {
	Message        int64
	Type           int //0=status,1=button,2=order
	ElevatorID     int
	TargetElevator int
}




type StatusMessage struct {
	Message        ElevatorStatus
	ElevatorID     int
	TargetElevator int
	MessageID      int64 //set by network module
}

type ButtonMessage struct {
	Message        int
	MessageType    bool //true for new ordre, false for completed
	ElevatorID     int
	TargetElevator int
	MessageID      int64 //set by network module
}

type OrderQueueNet struct {
	Elevator [MAX_ELEVATORS]map[string]bool
}

type OrderMessageNet struct {
	Message        OrderQueueNet
	ElevatorID     int
	TargetElevator int
	MessageID      int64 //set by network module
}

func NewOrderQueueNet() *OrderQueueNet {
	var orderQueueNet OrderQueueNet
	for i, _ := range orderQueueNet.Elevator {
		orderQueueNet.Elevator[i] = make(map[string]bool)
	}
	return &orderQueueNet
}

func NewOrderMessageNet() *OrderMessageNet {
	var orderMessageNet OrderMessageNet
	for i, _ := range orderMessageNet.Message.Elevator {
		orderMessageNet.Message.Elevator[i] = make(map[string]bool)
	}
	return &orderMessageNet
}




type ElevatorMovement struct {
	Dir         bool
	NextDir     bool
	TargetFloor int
}

type ElevatorStatus struct {
	Dir       bool
	LastFloor int
	Timeout   bool
	AtFloor   bool
	Idle      bool
	DoorOpen  bool
}

type OrderQueue struct {
	Elevator [MAX_ELEVATORS]map[int]bool
}

type OrderMessage struct {
	Message        OrderQueue
	ElevatorID     int
	TargetElevator int
	MessageID      int64 //set by network module
}

func NewOrderQueue() *OrderQueue {
	var orderQueue OrderQueue
	for i, _ := range orderQueue.Elevator {
		orderQueue.Elevator[i] = make(map[int]bool)
	}
	return &orderQueue
}

func NewOrderMessage() *OrderMessage {
	var orderMessage OrderMessage
	for i, _ := range orderMessage.Message.Elevator {
		orderMessage.Message.Elevator[i] = make(map[int]bool)
	}
	return &orderMessage
}

const (
	Init         int = 0
	Master       int = 1
	Slave        int = 2
	NoNetwork    int = 3
	DeadElevator int = 4
)

const EVERYONE = 0

var IPToID = map[string]int{
	//"129.241.187.144": 1, //labplass 12
	//"129.241.187.148": 2, //labplass 15
	//"129.241.187.142": 3, //labplass 14
	//"129.241.187.147": 3, //labplass 16
	"129.241.187.152": 3, //labplass 13
	//"129.241.187.157": 3,
}

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

const N_FLOORS = 4
const MAX_ELEVATORS = 3

var LightMatrix = [N_FLOORS][3]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var OrderButtonMatrix = [N_FLOORS][3]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

func GetButtonIndex(button int) (int, int) {
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < 3; j++ {
			if button == OrderButtonMatrix[i][j] {
				return i, j
			}
		}
	}
	return -1, -1
}

func GetLightIndex(light int) (int, int) {
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < 3; j++ {
			if light == LightMatrix[i][j] {
				return i, j
			}
		}
	}
	return -1, -1
}


func BtoS(button int) string {
	switch button {
	case BUTTON_UP1:
		return "UP1"
	case BUTTON_UP2:
		return "UP2"
	case BUTTON_UP3:
		return "UP3"
	case BUTTON_DOWN2:
		return "DOWN2"
	case BUTTON_DOWN3:
		return "DOWN3"
	case BUTTON_DOWN4:
		return "DOWN4"
	case BUTTON_COMMAND1:
		return "CDM1"
	case BUTTON_COMMAND2:
		return "CDM2"
	case BUTTON_COMMAND3:
		return "CDM3"
	case BUTTON_COMMAND4:
		return "CDM4"
	}
	return "Not a button"
}
