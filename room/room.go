package room

import (
	"github.com/googollee/go-socket.io"
	"math/rand"
	"sync"
	"github.com/kot13/gommo/monitor"
	"github.com/kot13/gommo/ticker"
	"time"
	"github.com/kot13/gommo/config"
)

const MAP_LOW_BOUND = 50
const MAP_HIGH_BOUND = 1950

type GameRoom struct {
	sync.RWMutex
	Name string
	Zoo *Zoo
	CommandMonitor *monitor.CommandMonitor
	sockets map[string]socketio.Socket
	worldTicker *ticker.WorldTicker
}

func NewGameRoom(name string, config config.RoomConfig, updateWorldAction func(room *GameRoom, socket socketio.Socket)) *GameRoom {
	gameRoom := &GameRoom {
		Name: name,
		Zoo: &Zoo {
			M: make(map[string]*Bunny),
		},
		CommandMonitor: monitor.NewCommandMonitor(config),
		sockets: make(map[string]socketio.Socket),
	}
	gameRoom.worldTicker = ticker.NewWorldTicker(func() {
		for _, socket := range gameRoom.sockets {
			updateWorldAction(gameRoom, socket)
		}
	}, time.Duration(config.RoomTickerPeriodMs) * time.Millisecond)
	return gameRoom
}

func (room *GameRoom) Connect(socket socketio.Socket) {
	room.updateConnection(func(room *GameRoom) {
		room.sockets[socket.Id()] = socket
	})
}

func (room *GameRoom) Disconnect(socket socketio.Socket) {
	room.updateConnection(func(room *GameRoom) {
		delete(room.sockets, socket.Id())
		delete(room.Zoo.M, socket.Id())
	})
}

func (room *GameRoom) updateConnection(action func(room *GameRoom)) {
	room.Lock()
	prevConnections := len(room.sockets)
	action(room)
	room.updateWorldTimer(prevConnections)
	room.Unlock()
}

func (room *GameRoom) updateWorldTimer(prevConnections int) {
	connections := len(room.sockets)
	if connections == 0 && prevConnections != 0 {
		room.worldTicker.Stop()
	} else if connections != 0 && prevConnections == 0 {
		room.worldTicker.Start()
	}
}

type Zoo struct {
	sync.RWMutex
	M map[string]*Bunny
}

func (zoo *Zoo) PlayerCount() int {
	zoo.Lock()
	playerCount := len(zoo.M)
	zoo.Unlock()
	return playerCount
}

type Bunny struct {
	Id      string `json:"id"`
	X       uint32 `json:"x"`
	Y       uint32 `json:"y"`
	Rotation float64 `json:"rotation"`
	Name    string `json:"name"`
	Width   uint32 `json:"wight"`
	Height  uint32 `json:"height"`
	IsAlive bool   `json:"isAlive"`
}

func NewBunny(id string, playerName string) *Bunny {
	return &Bunny{
		Id:      id,
		X:       uint32(rand.Intn(750)),
		Y:       uint32(rand.Intn(750)),
		Rotation: 0,
		Name:    playerName,
		Width:   32,
		Height:  32,
		IsAlive: true,
	}
}

func (bunny *Bunny) CheckBounds() {
	if bunny.X < MAP_LOW_BOUND { bunny.X = MAP_LOW_BOUND }
	if bunny.Y < MAP_LOW_BOUND { bunny.Y = MAP_LOW_BOUND }
	if bunny.X > MAP_HIGH_BOUND { bunny.X = MAP_HIGH_BOUND }
	if bunny.Y > MAP_HIGH_BOUND { bunny.Y = MAP_HIGH_BOUND }
}