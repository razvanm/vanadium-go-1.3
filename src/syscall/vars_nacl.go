// +build nacl

package syscall

import "sync"

var (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)
var ForkLock sync.RWMutex
var SocketDisableIPv6 bool
