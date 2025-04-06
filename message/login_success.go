package message

import "github.com/wspowell/tabletop/account"

type LoginSuccess struct {
	User account.User `json:"user"`
}

func (self LoginSuccess) Type() string {
	return "loginSuccess"
}
