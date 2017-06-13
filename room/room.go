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
	Name           string
	Battlefield    *Battlefield
	CommandMonitor *monitor.CommandMonitor
	sockets        map[string]socketio.Socket
	worldTicker    *ticker.WorldTicker
}

func NewGameRoom(name string, config config.RoomConfig, updateWorldAction func(room *GameRoom, socket socketio.Socket)) *GameRoom {
	gameRoom := &GameRoom{
		Name: name,
		Battlefield: &Battlefield{
			M: make(map[string]*Player),
		},
		CommandMonitor: monitor.NewCommandMonitor(config),
		sockets:        make(map[string]socketio.Socket),
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
		delete(room.Battlefield.M, socket.Id())
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

type Battlefield struct {
	sync.RWMutex
	M map[string]*Player
}

func (battlefield *Battlefield) PlayerCount() int {
	battlefield.Lock()
	playerCount := len(battlefield.M)
	battlefield.Unlock()
	return playerCount
}

type Player struct {
	Id       string `json:"id"`
	X        uint32 `json:"x"`
	Y        uint32 `json:"y"`
	Rotation float64 `json:"rotation"`
	Name     string `json:"name"`
	Width    uint32 `json:"wight"`
	Height   uint32 `json:"height"`
	IsAlive  bool   `json:"isAlive"`
}

func NewPlayer(id string, playerName string) *Player {
	return &Player{
		Id:       id,
		X:        uint32(rand.Intn(750)),
		Y:        uint32(rand.Intn(750)),
		Rotation: 0,
		Name:     playerName,
		Width:    32,
		Height:   32,
		IsAlive:  true,
	}
}

func (player *Player) CheckBounds() {
	if player.X < MAP_LOW_BOUND { player.X = MAP_LOW_BOUND }
	if player.Y < MAP_LOW_BOUND { player.Y = MAP_LOW_BOUND }
	if player.X > MAP_HIGH_BOUND { player.X = MAP_HIGH_BOUND }
	if player.Y > MAP_HIGH_BOUND { player.Y = MAP_HIGH_BOUND }
}