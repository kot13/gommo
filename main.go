package main

import (
	"encoding/json"
	"github.com/GeertJohan/go.rice"
	"github.com/googollee/go-socket.io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"fmt"
)

type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
}

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

func main() {
	http.Handle("/", http.FileServer(rice.MustFindBox("public").HTTPBox()))

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("On connection: " + so.Id())

		so.Join("room")

		so.On("join_new_player", func(playerName string) {
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

			sendToAll(so, "room", "add_players", string(bytes))
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
		so.On("player_move", func(location Location) {
			log.Println("Received location: " + fmt.Sprint(location))
			bytes, err := json.Marshal(map[string]string{
				"id": so.Id(),
				"x":  fmt.Sprint(location.X),
				"y":  fmt.Sprint(location.Y),
			})

			if err != nil {
				log.Println("Error marshal json")
			}

			sendToAll(so, "room", "player_position_update", string(bytes))
		})

		// ЕСЛИ КРОЛИК ВЫСТРЕЛИЛ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("shots_fired", func(id string) {
			sendToAll(so, "room", "player_fire_add", id)
		})

		// ЕСЛИ ПРОИЗОШЛО ПОПАДАНИЕ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_killed", func(victimId string) {
			zoo.Lock()
			zoo.m[victimId].IsAlive = false
			zoo.Unlock()

			sendToAll(so, "room", "clean_dead_player", victimId)
		})

		// ОПОВЕЩАЕМ О ДИСКОНЕКТЕ
		so.On("disconnection", func() {
			log.Println("On disconnect: " + so.Id())
			zoo.Lock()
			delete(zoo.m, so.Id())
			zoo.Unlock()
			so.BroadcastTo("room", "player_disconnect", so.Id())
		})
	})

	http.Handle("/ws/", server)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func sendToAll(so socketio.Socket, room, event string, args ...interface{}) {
	so.BroadcastTo(room, event, args)
	so.Emit(event, args)
}
