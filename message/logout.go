package message

type Logout struct {
	Username string `json:"username"`
}

func (self Logout) Type() string {
	return "logout"
}
