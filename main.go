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
	X        uint32 `json:"x"`
	Y        uint32 `json:"y"`
	Rotation float64 `json:"rotation"`
}

func NewPlayerLocation(player *room.Player) *PlayerLocation {
	return &PlayerLocation{
		X:        player.X,
		Y:        player.Y,
		Rotation: player.Rotation,
	}
}

type WorldStateForPlayer struct {
	Players  map[string]*room.Player `json:"players"`
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
			log.Println("Socket: " + so.Id() + ", NewPlayer: " + playerName)
			player := room.NewPlayer(so.Id(), playerName)
			gameRoom.Battlefield.Lock()
			gameRoom.Battlefield.M[so.Id()] = player
			gameRoom.Battlefield.Unlock()

		})

		so.On("player_rotation", func(rotation string) {
			gameRoom.Battlefield.Lock()
			gameRoom.Battlefield.M[so.Id()].Rotation, _ = strconv.ParseFloat(rotation, 64)
			gameRoom.Battlefield.Unlock()

			curTime := time.Now()
			gameRoom.CommandMonitor.Put(so.Id(), monitor.NewCommand("player_rotation", curTime, NewPlayerLocation(gameRoom.Battlefield.M[so.Id()])), curTime)
		})

		so.On("player_move", func(character string) {
			gameRoom.Battlefield.Lock()
			player := gameRoom.Battlefield.M[so.Id()]

			switch character {
			case "A": player.X -= 2
			case "S": player.Y += 2
			case "D": player.X += 2
			case "W": player.Y -= 2
			}

			player.CheckBounds()
			gameRoom.Battlefield.Unlock()

			curTime := time.Now()
			gameRoom.CommandMonitor.Put(so.Id(), monitor.NewCommand("player_move_" + character, curTime, NewPlayerLocation(gameRoom.Battlefield.M[so.Id()])), curTime)
		})

		so.On("shots_fired", func(id string) {
			//TODO need to inject bullets on server
			so.BroadcastTo("room", "player_fire_add", id)
			so.Emit("player_fire_add", id)
		})

		so.On("player_killed", func(victimId string) {
			gameRoom.Battlefield.Lock()
			gameRoom.Battlefield.M[victimId].IsAlive = false
			gameRoom.Battlefield.Unlock()
		})

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
		Players:  room.Battlefield.M,
		Commands: room.CommandMonitor.GetPlayerCommands(so.Id(), time.Now()),
	})
	room.Unlock()

	if err != nil {
		log.Println("Error marshal json")
	}

	so.Emit("world_update", string(bytes))
}