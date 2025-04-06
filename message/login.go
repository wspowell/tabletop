package message

type Login struct {
	Username string `json:"username"`
	Secret   string `json:"secret"`
}

func (self Login) Type() string {
	return "login"
}
