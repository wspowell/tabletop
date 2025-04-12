package message

type KeepAlive struct{}

func (self KeepAlive) Type() string {
	return "keepAlive"
}
