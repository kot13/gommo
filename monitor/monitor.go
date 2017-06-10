package monitor

import (
	"time"
	"sync"
	"github.com/kot13/gommo/config"
)

type Command struct {
	what string
	when time.Time
	result interface{}
}

func NewCommand(what string, when time.Time, result interface{}) *Command {
	return &Command{
		what: what,
		when: when,
		result: result,
	}
}

type CommandMonitor struct{
	sync.RWMutex
	stalePeriod time.Duration
	commandMap map[string][]*Command
}

func NewCommandMonitor(config config.RoomConfig) *CommandMonitor {
	return &CommandMonitor {
		stalePeriod: time.Duration(config.CommandStalePeriodMs) * time.Millisecond,
		commandMap: make(map[string][]*Command),
	}
}

func (monitor *CommandMonitor) Put(playerId string, command *Command, currentTime time.Time) {
	monitor.Lock()
	monitor.deleteStaleCommands(playerId, currentTime)
	monitor.commandMap[playerId] = append(monitor.commandMap[playerId], command)
	monitor.Unlock()
}

func (monitor *CommandMonitor) GetPlayerCommands(playerId string, currentTime time.Time) []*Command {
	monitor.Lock()
	defer monitor.Unlock()
	monitor.deleteStaleCommands(playerId, currentTime)
	return append([]*Command(nil), monitor.commandMap[playerId]...)
}

func (monitor *CommandMonitor) deleteStaleCommands(playerId string, currentTime time.Time) {
	commands, ok := monitor.commandMap[playerId]
	if ok {
		i := 0
		for _, command := range commands {
			if monitor.isCommandStale(currentTime, command) {
				commands[i] = command
				i++
			}
		}
		monitor.commandMap[playerId] = commands[:i]
	}
}

func (monitor *CommandMonitor) isCommandStale(currentTime time.Time, command *Command) bool {
	return currentTime.Sub(command.when) < monitor.stalePeriod
}
