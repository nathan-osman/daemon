package network

import (
	set "github.com/deckarep/golang-set"
	"log"
	"net"
	"time"
)

// Interval for refreshing the interfaces.
const refreshDuration = 10 * time.Second

// Monitor for the addition and removal of network interfaces.
type Watcher struct {
	Added   chan string
	Removed chan string
	Stop    chan bool
	names   set.Set
}

// Create a new watcher.
func NewWatcher() *Watcher {
	w := &Watcher{
		Added:   make(chan string),
		Removed: make(chan string),
		Stop:    make(chan bool),
		names:   set.NewSet(),
	}
	go w.watch()
	return w
}

// Refresh the list of interfaces.
func (w *Watcher) refresh() error {
	names := set.NewSet()
	ifis, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, ifi := range ifis {
		if ifi.Flags&net.FlagUp != 0 && ifi.Flags&(net.FlagBroadcast|net.FlagMulticast) != 0 {
			names.Add(ifi.Name)
		}
	}
	for name := range names.Difference(w.names).Iter() {
		w.Added <- name.(string)
	}
	for name := range w.names.Difference(names).Iter() {
		w.Removed <- name.(string)
	}
	w.names = names
	return nil
}

// Continuously watch for interface changes.
func (w *Watcher) watch() {
	w.refresh()
	t := time.NewTicker(refreshDuration)
	for {
		select {
		case <-t.C:
			if err := w.refresh(); err != nil {
				log.Println(err)
			}
		case <-w.Stop:
			return
		}
	}
}
