package message

type MapList struct {
	MapNames []string `json:"mapNames"`
}

func (self MapList) Type() string {
	return "mapList"
}
