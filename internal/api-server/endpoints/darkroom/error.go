package darkroom

type Error struct {
	Message string `json:"message"`
	Err     string `json:"error"`
	Code    int    `json:"-"`
}

func (e Error) Error() string {
	return e.Message
}
