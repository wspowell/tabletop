package message

type UserOffline struct {
	Username string `json:"username"`
}

func (self UserOffline) Type() string {
	return "userOffline"
}
