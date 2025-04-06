package message

type DeleteToken struct {
	Id      string `json:"id"`
	MapName string `json:"mapName"`
}

func (self DeleteToken) Type() string {
	return "deleteToken"
}
