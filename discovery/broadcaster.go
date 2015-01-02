package discovery

import (
	"log"
	"net"
	"time"
)

// Standard port for broadcasting.
const broadcastPort = 17200

// Interval for broadcasting packets.
const broadcastDuration = 10 * time.Second

// Broadcasts on available network interfaces.
type Broadcaster struct {
	watcher     *watcher
	connections map[string]*net.UDPConn
}

// Create a new broadcaster.
func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		watcher:     NewWatcher(),
		connections: make(map[string]*net.UDPConn),
	}
	go b.broadcast()
	return b
}

// Create a connection to an interface's broadcast address.
func (b *Broadcaster) connect(name string) (*net.UDPConn, error) {
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	if ifi.Flags&net.FlagBroadcast != 0 {
		return net.DialUDP("udp4", nil, &net.UDPAddr{
			IP:   net.ParseIP("255.255.255.255"),
			Port: broadcastPort,
		})
	} else {
		return net.DialUDP("udp6", nil, &net.UDPAddr{
			IP:   net.ParseIP("ff02::1"),
			Port: broadcastPort,
		})
	}
}

// Broadcast on all interfaces.
func (b *Broadcaster) broadcast() {
	c := time.Tick(broadcastDuration)
	for {
		select {
		case <-c:
			for _, conn := range b.connections {
				conn.Write([]byte("test"))
			}
		case i := <-b.watcher.InterfaceAdded:
			if conn, err := b.connect(i); err != nil {
				log.Println(err)
			} else {
				b.connections[i] = conn
			}
		case i := <-b.watcher.InterfaceRemoved:
			delete(b.connections, i)
		}
	}
}
