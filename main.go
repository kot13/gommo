package main

import (
	"encoding/json"
	"github.com/GeertJohan/go.rice"
	"github.com/googollee/go-socket.io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Zoo struct {
	sync.RWMutex
	m map[string]*Bunny
}

type Bunny struct {
	Id      string `json:"id"`
	X       uint32 `json:"x"`
	Y       uint32 `json:"y"`
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
		Name:    playerName,
		Width:   32,
		Height:  32,
		IsAlive: true,
	}
}

var zoo Zoo = Zoo{
	m: make(map[string]*Bunny),
}

const MAP_LOW_BOUND = 50
const MAP_HIGH_BOUND = 1950

var worldTimer *WorldTimer

func main() {
	http.Handle("/", http.FileServer(rice.MustFindBox("public").HTTPBox()))

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("On connection: " + so.Id())

		worldTimer = NewWorldTimer(so, sendSnapshot, 40 * time.Millisecond)
		so.Join("room")

		so.On("join_new_player", func(playerName string) {
			prevPlayerCount := zoo.playerCount()

			// СОЗДАЁМ КРОЛИКА
			bunny := NewBunny(so.Id(), playerName)
			zoo.Lock()
			zoo.m[so.Id()] = bunny
			zoo.Unlock()

			// ОТПРАВЛЯЕМ ЗООПАРК НА КЛИЕНТА
			zoo.Lock()
			bytes, err := json.Marshal(zoo.m)
			zoo.Unlock()
			if err != nil {
				log.Println("Error marshal json")
			}

			updateWorldTimer(prevPlayerCount)

			so.BroadcastTo("room", "add_players", string(bytes))
			so.Emit("add_players", string(bytes))
		})

		// ЕСЛИ КРОЛИК ПОВЕРНУЛСЯ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_rotation", func(rotation string) {
			bytes, err := json.Marshal(map[string]string{
				"id":       so.Id(),
				"rotation": rotation,
			})
			if err != nil {
				log.Println("Error marshal json")
			}
			so.BroadcastTo("room", "player_rotation_update", string(bytes))
		})

		// ЕСЛИ КРОЛИК ДВИГАЕТСЯ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_move", func(character string) {
			zoo.Lock()
			bunny := zoo.m[so.Id()]

			switch character {
			case "A":
				bunny.X -= 2
			case "S":
				bunny.Y += 2
			case "D":
				bunny.X += 2
			case "W":
				bunny.Y -= 2
			}

			bunny.checkBounds()
			zoo.Unlock()
		})

		// ЕСЛИ КРОЛИК ВЫСТРЕЛИЛ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("shots_fired", func(id string) {
			so.BroadcastTo("room", "player_fire_add", id)
			so.Emit("player_fire_add", id)
		})

		// ЕСЛИ ПРОИЗОШЛО ПОПАДАНИЕ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_killed", func(victimId string) {
			zoo.Lock()
			zoo.m[victimId].IsAlive = false
			zoo.Unlock()

			so.BroadcastTo("room", "clean_dead_player", victimId)
			so.Emit("clean_dead_player", victimId)
		})

		// ОПОВЕЩАЕМ О ДИСКОНЕКТЕ
		so.On("disconnection", func() {
			prevPlayerCount := zoo.playerCount()

			log.Println("On disconnect: " + so.Id())
			zoo.Lock()
			delete(zoo.m, so.Id())
			zoo.Unlock()

			updateWorldTimer(prevPlayerCount)

			so.BroadcastTo("room", "player_disconnect", so.Id())
		})
	})

	http.Handle("/ws/", server)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (bunny *Bunny) checkBounds() {
	if bunny.X < MAP_LOW_BOUND { bunny.X = MAP_LOW_BOUND }
	if bunny.Y < MAP_LOW_BOUND { bunny.Y = MAP_LOW_BOUND }
	if bunny.X > MAP_HIGH_BOUND { bunny.X = MAP_HIGH_BOUND }
	if bunny.Y > MAP_HIGH_BOUND { bunny.Y = MAP_HIGH_BOUND }
}

func (zoo *Zoo) playerCount() int {
	zoo.Lock()
	playerCount := len(zoo.m)
	zoo.Unlock()
	return playerCount
}

func updateWorldTimer(prevPlayerCount int) {
	playerCount := zoo.playerCount()
	if playerCount == 0 && prevPlayerCount != 0 {
		worldTimer.Stop()
	} else if playerCount != 0 && prevPlayerCount == 0 {
		worldTimer.Start()
	}
}

func sendSnapshot(so socketio.Socket) {
	sendToAll(so, "room", "world_update", zoo.m)
}

func sendToAll(so socketio.Socket, room, event string, args ...interface{}) {
	so.BroadcastTo(room, event, args...)
	so.Emit(event, args...)
}