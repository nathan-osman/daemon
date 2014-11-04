package network

import (
	set "github.com/deckarep/golang-set"
	"log"
	"net"
	"time"
)

// Interval for refreshing the interfaces.
const refreshDuration = 10 * time.Second

// Status update on an interface.
type Status struct {
	Status bool
	Name   string
}

// Watcher for the addition and removal of network interfaces.
type Watcher struct {
	Stop      chan bool
	listeners []chan Status
	oldNames  set.Set
}

// Create a new watcher.
func NewWatcher() *Watcher {
	w := &Watcher{
		Stop:     make(chan bool, 1),
		oldNames: set.NewSet(),
	}
	go w.watch()
	return w
}

// Return a channel for receiving notifications.
func (w *Watcher) Listen() chan Status {
	c := make(chan Status)
	w.listeners = append(w.listeners, c)
	return c
}

// Send a message to all listeners.
func (w *Watcher) notify(status Status) {
	for _, c := range w.listeners {
		c <- status
	}
}

// Refresh the list of interfaces.
func (w *Watcher) refresh() error {
	newNames := set.NewSet()
	ifis, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, ifi := range ifis {
		if ifi.Flags&net.FlagUp != 0 && ifi.Flags&(net.FlagBroadcast|net.FlagMulticast) != 0 {
			newNames.Add(ifi.Name)
		}
	}
	for name := range newNames.Difference(w.oldNames).Iter() {
		w.notify(Status{true, name.(string)})
	}
	for name := range w.oldNames.Difference(newNames).Iter() {
		w.notify(Status{false, name.(string)})
	}
	w.oldNames = newNames
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
