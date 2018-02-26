package messagingapi

import (
	"io"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
)

const (
	DIRECT 		= "direct"
	BROADCAST 	= "broadcast"
)


type Contact struct {
	Id					int64			`json:"id"`
	Sender 				*Principal 		`json:"sender"`
	Receiver 			*Principal 		`json:"receiver"`
	ContactType			string			`json:"type"`
}

func ContactFromJson(data io.Reader) *Contact {
	decoder := json.NewDecoder(data)
	var contact Contact
	err := decoder.Decode(&contact)
	if err == nil {
		return &contact
	} else {
		l4g.Error(err.Error())
		return nil
	}
}



func (c *Contact) ToJson() string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
