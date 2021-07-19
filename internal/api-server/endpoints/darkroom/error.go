package darkroom

type Error struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func (e Error) Error() string {
	return e.Message
}
