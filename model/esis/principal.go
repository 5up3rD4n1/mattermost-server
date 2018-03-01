package esis

import "fmt"

const (
	USER  = "USR"
	GROUP = "GRP"
)

type Principal struct {
	Id            int64  `json:"id"`
	Handle        string `json:"handle"`
	PrincipalType string `json:"type"`
	AsUser        *User	 `json:"asUser"`
	AsUserGroup	  *Group `json:"asUserGroup"`
}

type Group struct {
	Name      string
	Principal *Principal
	ApiUsers  *[]User
}

type TimeWindow struct {
	Hour 	int64 `json:"hour"`
	Minute 	int64 `json:"minute"`
	Second 	int64 `json:"second"`
	Nano 	int64 `json:"nano"`
}

type User struct {
	Id		  			int64 		`json:"id"`
	Username  			string 		`json:"username"`
	Password  			string		`json:"password"`
	FullName  			string		`json:"fullName"`
	Principal 			int64 		`json:"principal"`
	ReceiptWindowStart 	*TimeWindow `json:"receiptWindowStart"`
	ReceiptWindowEnd 	*TimeWindow `json:"receiptWindowEnd"`
}

func (t *TimeWindow) String() string {
	if t == nil {
		return ""
	}

	return fmt.Sprintf("%v:%v:%v:%v",t.Hour, t.Minute, t.Second, t.Nano)
}
