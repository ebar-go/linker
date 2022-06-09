package linker

import (
	"errors"
)

type GroupServer struct {
	items []IServer
}

func (group *GroupServer) Use(handlers ...HandleFunc) {
	for _, item := range group.items {
		item.Use(handlers...)
	}
}

func (group *GroupServer) Register(server ...IServer) {
	group.items = append(group.items, server...)
}

func (group *GroupServer) Start() error {
	if len(group.items) == 0 {
		return errors.New("no server register")
	}

	for _, item := range group.items {
		if err := item.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (group *GroupServer) SetOnConnect(f func(connection IConnection)) {
	for _, item := range group.items {
		item.SetOnConnect(f)
	}
}

func (group *GroupServer) SetOnDisconnect(f func(connection IConnection)) {
	for _, item := range group.items {
		item.SetOnDisconnect(f)
	}
}

func (group *GroupServer) SetOnReceive(hookFunc HandleFunc) {
	for _, item := range group.items {
		item.SetOnReceive(hookFunc)
	}
}
