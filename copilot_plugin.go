package main

type UserRequest struct {
	Message string
}

type ChunkData struct {
	ID string
	Content string
	IsLast bool
}

type CopilotPlugin interface {
	Chat(UserRequest) (<-chan ChunkData, error)

	AutoComplete(UserRequest) (string, error)
}
