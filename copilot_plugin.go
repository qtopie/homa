package main

import "io"

type UserRequest struct {
	Message string
}

type CopilotPlugin interface {
	Chat(UserRequest) (io.WriteCloser, error)

	AutoComplete(UserRequest) (string, error)
}
