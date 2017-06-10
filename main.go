package main

import (
	"encoding/json"
	"net/http"
	"time"
	"strconv"

	"github.com/GeertJohan/go.rice"
	"github.com/googollee/go-socket.io"
	"github.com/kot13/gommo/config"
	"github.com/kot13/gommo/logger"
	"github.com/kot13/gommo/room"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"github.com/kot13/gommo/monitor"
)

type PlayerLocation struct {
	X uint32 `json:"x"`
	Y uint32 `json:"y"`
	Rotation float64 `json:"rotation"`
}

func NewPlayerLocation(bunny *room.Bunny) *PlayerLocation {
	return &PlayerLocation{
		X: bunny.X,
		Y: bunny.Y,
		Rotation: bunny.Rotation,
	}
}

type WorldStateForPlayer struct {
	Players map[string]*room.Bunny `json:"players"`
	Commands []*monitor.Command `json:"commands"`
}

func main() {
	conf := config.GetConfig()
	logFinalizer, err := logger.InitLogger(conf.Logger)
	if err != nil {
		log.Fatal(err)
	}
	defer logFinalizer()

	log.Info("start")

	gameRoom := room.NewGameRoom("room", conf.Room, sendSnapshot)
	http.Handle("/", http.FileServer(rice.MustFindBox("public").HTTPBox()))

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("On connection: " + so.Id() + " - " + fmt.Sprint(&so))

		so.Join(gameRoom.Name)
		gameRoom.Connect(so)

		so.On("join_new_player", func(playerName string) {
			// СОЗДАЁМ КРОЛИКА
			log.Println("Socket: " + so.Id() + ", NewPlayer: " + playerName)
			bunny := room.NewBunny(so.Id(), playerName)
			gameRoom.Zoo.Lock()
			gameRoom.Zoo.M[so.Id()] = bunny
			gameRoom.Zoo.Unlock()

		})

		// ЕСЛИ КРОЛИК ПОВЕРНУЛСЯ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_rotation", func(rotation string) {
			gameRoom.Zoo.Lock()
			gameRoom.Zoo.M[so.Id()].Rotation, _ = strconv.ParseFloat(rotation, 64)
			gameRoom.Zoo.Unlock()

			curTime := time.Now()
			gameRoom.CommandMonitor.Put(so.Id(), monitor.NewCommand("player_rotation", curTime, NewPlayerLocation(gameRoom.Zoo.M[so.Id()])), curTime)
		})

		// ЕСЛИ КРОЛИК ДВИГАЕТСЯ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_move", func(character string) {
			gameRoom.Zoo.Lock()
			bunny := gameRoom.Zoo.M[so.Id()]

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

			bunny.CheckBounds()
			gameRoom.Zoo.Unlock()

			curTime := time.Now()
			gameRoom.CommandMonitor.Put(so.Id(), monitor.NewCommand("player_move_" + character, curTime, NewPlayerLocation(gameRoom.Zoo.M[so.Id()])), curTime)
		})

		// ЕСЛИ КРОЛИК ВЫСТРЕЛИЛ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("shots_fired", func(id string) {
			//TODO need to inject bullets on server
			so.BroadcastTo("room", "player_fire_add", id)
			so.Emit("player_fire_add", id)
		})

		// ЕСЛИ ПРОИЗОШЛО ПОПАДАНИЕ ОПОВЕЩАЕМ КЛИЕНТОВ
		so.On("player_killed", func(victimId string) {
			gameRoom.Zoo.Lock()
			gameRoom.Zoo.M[victimId].IsAlive = false
			gameRoom.Zoo.Unlock()
		})

		// ОПОВЕЩАЕМ О ДИСКОНЕКТЕ
		so.On("disconnection", func() {
			log.Println("On disconnect: " + so.Id())
			gameRoom.Disconnect(so)
		})
	})

	http.Handle("/ws/", server)

	err = http.ListenAndServe(conf.App.AppPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func sendSnapshot(room *room.GameRoom, so socketio.Socket) {
	room.Lock()
	bytes, err := json.Marshal(&WorldStateForPlayer{
		Players: room.Zoo.M,
		Commands: room.CommandMonitor.GetPlayerCommands(so.Id(), time.Now()),
	})
	room.Unlock()

	if err != nil {
		log.Println("Error marshal json")
	}

	so.Emit("world_update", string(bytes))
}