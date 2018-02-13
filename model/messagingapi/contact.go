package messagingapi

import (
	"io"
	"encoding/json"
)

const (
	DIRECT 		= "direct"
	BROADCAST 	= "broadcast"
)


type Contact struct {
	Id					int64			`json:"id"`
	// Sender 				*ApiPrincipal 	`json:"sender,omitempty"` Omit field due to API ERROR
	Receiver 			*Principal 	`json:"receiver"`
	ContactType			string			`json:"type"`
	ReceiptWindowStart	string			`json:"receiptWindowStart"`
	ReceiptWindowEnd	string			`json:"receiptWindowEnd"`
}

func ContactFromJson(data io.Reader) *Contact {
	decoder := json.NewDecoder(data)
	var o Contact
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}
