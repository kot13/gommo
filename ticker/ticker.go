package ticker

import (
	"time"
	"github.com/googollee/go-socket.io"
)

type WorldTicker struct {
	socket         socketio.Socket
	action         func()
	updateDuration time.Duration

	performer *time.Ticker
	quit      chan struct{}
}

func NewWorldTicker(whatFunc func(), updateDuration time.Duration) *WorldTicker {
	return &WorldTicker{
		updateDuration: updateDuration,
		action:         whatFunc,
	}
}

func (ticker *WorldTicker) Start() {
	ticker.performer = time.NewTicker(ticker.updateDuration)
	ticker.quit = make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.performer.C:
				ticker.action()
			case <-ticker.quit:
				ticker.performer.Stop()
				return
			}
		}
	}()
}

func (ticker *WorldTicker) Stop() {
	close(ticker.quit)
}