package shared

type UserRequest struct {
	SessionId string
	Seq       int32
	Message   string
	FrontPart string
	BackPart string
	Filename  string
	Workspace string
}

type ChunkData struct {
	ID      string
	Content string
	IsLast  bool
}
