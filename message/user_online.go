package message

import "github.com/wspowell/tabletop/account"

type UserOnline struct {
	User account.User `json:"user"`
}

func (self UserOnline) Type() string {
	return "userOnline"
}
