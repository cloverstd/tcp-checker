package tcpchecker

import (
	"time"
)

// Option is a tcp checker option
type Option struct {
	Interval    time.Duration
	Fall        int
	Rise        int
	Timeout     time.Duration
	DefaultDown bool
}
