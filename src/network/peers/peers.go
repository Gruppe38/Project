package peers

import (
	. "../../defs/"
	"../conn"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"
)

const interval = 15 * time.Millisecond
const timeout = 300 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := <-transmitEnable
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}

func EstablishConnection(peerUpdateCh <-chan PeerUpdate, peerTxEnable chan<- bool, masterIDUpdate chan<- int,
	masterBroadcast <-chan PeerUpdate, masterBroadcastEnable chan<- bool, myID int, state *int) {
	timer := time.NewTimer(45 * time.Millisecond)
	numberOfPeers := 0
	select {
	case p := <-peerUpdateCh:
		numberOfPeers = len(p.Peers)
	case <-timer.C:
		break
	}
	peerTxEnable <- true

	masterID := -1
	if numberOfPeers == 0 {
		fmt.Println("I am master", myID)
		masterID = myID
		*state = Master
	} else {
		m := <-masterBroadcast
		masterID, _ = strconv.Atoi(m.Peers[0])
		fmt.Println("I am not master, master is", masterID)
		*state = Slave
	}
	masterIDUpdate <- masterID
}
