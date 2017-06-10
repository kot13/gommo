package monitor

import (
	"time"
	"sync"
)

const STALE_PERIOD = time.Second

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

func (command *Command) isStale(currentTime time.Time) bool {
	return currentTime.Sub(command.when) >= STALE_PERIOD
}

type CommandMonitor struct{
	sync.RWMutex
	commandMap map[string][]*Command
}

func NewCommandMonitor() *CommandMonitor {
	return &CommandMonitor {
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
			if !command.isStale(currentTime) {
				commands[i] = command
				i++
			}
		}
		monitor.commandMap[playerId] = commands[:i]
	}
}
