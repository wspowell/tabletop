package message

type TokenBank struct {
	TokenNames []string `json:"tokenNames"`
}

func (self TokenBank) Type() string {
	return "tokenBank"
}
