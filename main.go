package main

import (
	"net/http"
	"log"
	"github.com/GeertJohan/go.rice"
	"github.com/googollee/go-socket.io"
	"encoding/json"
	"sync"
	"math/rand"
)

type Zoo struct {
	sync.RWMutex
	m map[string]*Bunny
}

type Bunny struct {
	Id string `json:"id"`
	X uint32 `json:"x"`
	Y uint32 `json:"y"`
	Name string `json:"name"`
	Width uint32 `json:"wight"`
	Height uint32 `json:"height"`
	IsAlive bool `json:"isAlive"`
}

func NewBunny(id string) *Bunny{
	return &Bunny{
		Id: id,
		X: uint32(rand.Intn(750)),
		Y: uint32(rand.Intn(750)),
		Name: "",
		Width: 32,
		Height: 32,
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

		// СОЗДАЁМ КРОЛИКА
		bunny := NewBunny(so.Id())
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
		so.BroadcastTo("room", "add_players", string(bytes))
		so.Emit("add_players", string(bytes))

		// ЕСЛИ КРОЛИК ПОВЕРНУЛСЯ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_rotation", func(rotation string) {
			bytes, err = json.Marshal(map[string]string{
				"id": so.Id(),
				"rotation": rotation,
			})
			if err != nil {
				log.Println("Error marshal json")
			}
			so.BroadcastTo("room", "player_rotation_update", string(bytes))
		})

		// ЕСЛИ КРОЛИК ДВИГАЕТСЯ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_move", func(character string) {
			x := "0"
			y := "0"

			switch character {
			case "A":
				x = "-5"
			case "S":
				y = "5"
			case "D":
				x = "5"
			case "W":
				y = "-5"
			}

			bytes, err = json.Marshal(map[string]string{
				"id": so.Id(),
				"x": x,
				"y": y,
			})
			if err != nil {
				log.Println("Error marshal json")
			}

			so.BroadcastTo("room", "player_position_update", string(bytes))
			so.Emit("player_position_update", string(bytes))
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
			log.Println("On disconnect: " + so.Id())
			zoo.Lock()
			delete(zoo.m, so.Id())
			zoo.Unlock()
			so.BroadcastTo("room", "player_disconnect", string(bytes))
		})
	})

	http.Handle("/ws/", server)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}