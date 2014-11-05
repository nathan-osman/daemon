package discovery

import (
	set "github.com/deckarep/golang-set"
	"github.com/nitroshare/daemon/util"
	"log"
	"net"
	"time"
)

// Interval for refreshing the interfaces.
const refreshDuration = 10 * time.Second

// Status update on an interface.
type status struct {
	Status bool
	Name   string
}

// Watcher for the addition and removal of network interfaces.
type watcher struct {
	StatusChanged *util.Signal
	oldNames      set.Set
}

// Create a new watcher.
func NewWatcher() *watcher {
	w := &watcher{
		StatusChanged: &util.Signal{},
		oldNames:      set.NewSet(),
	}
	go w.watch()
	return w
}

// Refresh the list of interfaces.
func (w *watcher) refresh() error {
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
		w.StatusChanged.Emit(status{true, name.(string)})
	}
	for name := range w.oldNames.Difference(newNames).Iter() {
		w.StatusChanged.Emit(status{false, name.(string)})
	}
	w.oldNames = newNames
	return nil
}

// Continuously watch for interface changes.
func (w *watcher) watch() {
	w.refresh()
	c := time.Tick(refreshDuration)
	for _ = range c {
		if err := w.refresh(); err != nil {
			log.Println(err)
		}
	}
}
