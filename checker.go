package tcpchecker

import (
	"sync"
	"time"

	"github.com/tevino/abool"
	tcp "github.com/tevino/tcp-shaker"
)

const (
	defaultCheckInterval = time.Millisecond * 500
	defaultcheckTimeout  = time.Millisecond * 200
	defaultFail          = 3
	defaultRise          = 2
)

// Host is a host status wrapper.
type Host struct {
	addr string

	down *abool.AtomicBool

	checker *Checker

	refCountMutex sync.Mutex
	refCount      int

	stopOnce sync.Once
	stopC    chan struct{}
}

func (host *Host) worker() {
	timer := time.NewTimer(host.checker.option.Interval)
	defer timer.Stop()
	var (
		fall = host.checker.option.Fall
		rise = host.checker.option.Rise
	)
	check := func() {
		err := host.checker.shaker.CheckAddr(host.addr, host.checker.option.Timeout)
		if err != nil {
			fall--
			if fall == 0 {
				host.down.Set()
				fall = host.checker.option.Fall
				rise = host.checker.option.Rise
			}
		} else {
			rise--
			if rise == 0 {
				host.down.UnSet()
				rise = host.checker.option.Rise
				fall = host.checker.option.Fall
			}
		}
	}
	check()
	for {
		select {
		case <-timer.C:
			check()
			timer.Reset(host.checker.option.Interval)
		case <-host.stopC:
			return
		}
	}
}

func (host *Host) stop() {
	host.stopOnce.Do(func() {
		close(host.stopC)
	})
}

// Checker is a TCPChecker.
type Checker struct {
	sync.RWMutex
	bucket map[string]*Host
	shaker *tcp.Checker

	option Option
}

// Down returened the status of addr.
func (checker *Checker) Down(addr string) bool {
	checker.RLock()
	host, exist := checker.bucket[addr]
	if !exist {
		checker.RUnlock()
		checker.AddRef(addr)
		return checker.option.DefaultDown
	}
	checker.RUnlock()
	return host.down.IsSet()
}

// AddRef to checker.
func (checker *Checker) AddRef(addr string) {
	checker.RLock()
	host, exist := checker.bucket[addr]
	if exist {
		host.refCountMutex.Lock()
		host.refCount++
		host.refCountMutex.Unlock()
		checker.RUnlock()
		return
	}
	checker.RUnlock()
	checker.Lock()
	host, exist = checker.bucket[addr]
	if exist {
		host.refCountMutex.Lock()
		host.refCount++
		host.refCountMutex.Unlock()
		checker.Unlock()
		return
	}
	host = &Host{
		addr:     addr,
		refCount: 1,
		stopC:    make(chan struct{}),
		checker:  checker,
		down:     abool.New(),
	}
	host.down.SetTo(checker.option.DefaultDown)
	checker.bucket[addr] = host
	go host.worker()
	checker.Unlock()

}

// UnRef the addr from checker.
func (checker *Checker) UnRef(addr string) {
	checker.RLock()
	host, exist := checker.bucket[addr]
	if exist {
		host.refCountMutex.Lock()
		host.refCount--
		if host.refCount == 0 {
			host.stop()
			delete(checker.bucket, addr)
		}
		host.refCountMutex.Unlock()
		return
	}
	checker.RUnlock()
}

// New returned the Checker with option.
func New(option Option) (*Checker, error) {
	if option.Fall <= 0 {
		option.Fall = defaultFail
	}
	if option.Rise <= 0 {
		option.Rise = defaultRise
	}
	if option.Interval <= 0 {
		option.Interval = defaultCheckInterval
	}
	if option.Timeout <= 0 {
		option.Timeout = defaultcheckTimeout
	}
	shaker := tcp.NewChecker(true)
	err := shaker.InitChecker()
	if err != nil {
		return nil, err
	}
	checker := &Checker{
		option: option,
		bucket: make(map[string]*Host),
		shaker: shaker,
	}
	return checker, nil
}
