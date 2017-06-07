package main

import (
	"time"
	"sync"
)

type WorldTimer struct {
	sync.RWMutex

	whatFunc func()
	updateDuration time.Duration

	ticker *time.Ticker
	quit chan struct{}
}

func NewWorldTimer(whatFunc func(), updateDuration time.Duration) *WorldTimer {
	return &WorldTimer{
		updateDuration: updateDuration,
		whatFunc: whatFunc,
	}
}

func (timer *WorldTimer) Start() {
	timer.Lock()
	timer.ticker = time.NewTicker(40 * time.Millisecond)
	timer.quit = make(chan struct{})
	timer.Unlock()

	go func() {
		for {
			select {
			case <- timer.ticker.C:
				timer.whatFunc()
			case <- timer.quit:
				timer.ticker.Stop()
				return
			}
		}
	}()
}

func (timer *WorldTimer) Stop() {
	timer.Lock()
	close(timer.quit)
	timer.Unlock()
}