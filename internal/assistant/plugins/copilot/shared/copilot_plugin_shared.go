package shared

type UserRequest struct {
	Message string
}

type ChunkData struct {
	ID string
	Content string
	IsLast bool
}
