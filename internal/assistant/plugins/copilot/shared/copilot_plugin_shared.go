package shared

type UserRequest struct {
	SessionId string `json:"-"`
	Seq       int32 `json:"-"`
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
