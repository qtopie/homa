package main

import "github.com/qtopie/homa/internal/assistant/plugins/copilot/shared"

type CopilotPlugin interface {
	Chat(shared.UserRequest) (<-chan shared.ChunkData, error)

	AutoComplete(shared.UserRequest) (string, error)
}