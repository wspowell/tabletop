package message

type Error struct {
	TypeOfFailedMessage string `json:"typeOfFailedMessage"`
	ErrorMessage        string `json:"errorMessage"`
}

func (self Error) Type() string {
	return "error"
}
