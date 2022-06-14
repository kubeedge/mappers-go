package common

import (
	"sync"
)

// Lock the mutex lock of device, in the future, ddd some control information to
// limit the time for each device property to obtain resources
type Lock struct {
	DeviceLock *sync.Mutex
}

// Lock device get lock
func (dl *Lock) Lock() {
	dl.DeviceLock.Lock()
}

// Unlock device release lock
func (dl *Lock) Unlock() {
	dl.DeviceLock.Unlock()
}
