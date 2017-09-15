package channelsync

import (
	"sync"
)

// sync.Locker

type Locker sync.Locker

// sync.Mutex

type Mutex struct {
	c chan struct{}
}

func NewMutex() *Mutex {
	return &Mutex{
		c: make(chan struct{}, 1),
	}
}

func (cm *Mutex) Lock() {
	cm.c <- struct{}{}
}

func (cm *Mutex) Unlock() {
	<-cm.c
}

// sync.RWMutex

type RWMutex struct {
	c   chan struct{} // main mutex
	rn  int           // number of current readers
	rnc chan struct{} // mutex for rn
}

func NewRWMutex() *RWMutex {
	return &RWMutex{
		c:   make(chan struct{}, 1),
		rnc: make(chan struct{}, 1),
	}
}

func (cm *RWMutex) Lock() {
	cm.c <- struct{}{}
}

func (cm *RWMutex) Unlock() {
	<-cm.c
}

func (cm *RWMutex) RLock() {
	cm.rnc <- struct{}{}
	if cm.rn == 0 {
		cm.c <- struct{}{}
	}
	cm.rn++
	<-cm.rnc
}

func (cm *RWMutex) RUnlock() {
	cm.rnc <- struct{}{}
	cm.rn--
	if cm.rn == 0 {
		<-cm.c
	}
	<-cm.rnc
}

func (cm *RWMutex) RLocker() Locker {
	return (*rlocker)(cm)
}

type rlocker RWMutex

func (r *rlocker) Lock()   { (*RWMutex)(r).RLock() }

func (r *rlocker) Unlock() { (*RWMutex)(r).RUnlock() }

// sync.Once

type Once struct {
	c    chan struct{}
	done chan struct{}
}

func NewOnce() *Once {
	return &Once{
		c:    make(chan struct{}, 1),
		done: make(chan struct{}),
	}
}

func (co *Once) Do(f func()) {
	select {
	case co.c <- struct{}{}:
		f()
		close(co.done)
	case <-co.done:
	}
}

// sync.WaitGroup

type WaitGroup struct {
	n    int64
	lock chan struct{}
	wait chan struct{}
}

func NewWaitGroup() *WaitGroup {
	cwg := &WaitGroup{
		lock: make(chan struct{}, 1),
		wait: make(chan struct{}),
	}
	close(cwg.wait)
	return cwg
}

func (cwg *WaitGroup) Add(delta int) {
	cwg.lock <- struct{}{}
	cwg.n += int64(delta)
	if delta > 0 {
		if cwg.n == int64(delta) {
			cwg.wait = make(chan struct{})
		}
	} else if cwg.n == 0 {
		if delta < 0 {
			close(cwg.wait)
		}
	} else if cwg.n < 0 {
		panic("sync: negative WaitGroup counter")
	}
	<-cwg.lock
}

func (cwg *WaitGroup) Done() {
	cwg.Add(-1)
}

func (cwg *WaitGroup) Wait() {
	cwg.lock <- struct{}{}
	wait := cwg.wait
	<-cwg.lock
	<-wait
}

// sync.Cond

type Cond struct {
	L Locker
	c chan struct{}
}

func NewCond(l Locker) *Cond {
	return &Cond{
		L: l,
		c: make(chan struct{}),
	}
}

func (cc *Cond) Wait() {
	cc.L.Unlock()
	<-cc.c
	cc.L.Lock()
}

func (cc *Cond) Signal() {
	cc.c <- struct{}{}
}

func (cc *Cond) Broadcast() {
	for {
		select {
		case cc.c <- struct{}{}:
		default:
			return
		}
	}
}

// sync.Pool (hard to reimplement it)

type Pool sync.Pool
