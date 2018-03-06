package camino

import "fmt"

const (
	PRINCIPAL_USER  = "USR"
	PRINCIPAL_GROUP = "GRP"
)

type Principal struct {
	Id            int64  `json:"id"`
	Handle        string `json:"handle"`
	Type 		  string `json:"type"`
	AsUser        *User	 `json:"asUser"`
	AsUserGroup	  *Group `json:"asUserGroup"`
}

type Group struct {
	Name      string
	Principal *Principal
	Users  []*User
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

type Tenant struct {
	Id 			int64  `json:"id"`
	Name 		string `json:"name"`
	DisplayName string `json:"displayName"`
	Handle 		string `json:"handle"`
}

func (t *TimeWindow) String() string {
	if t == nil {
		return ""
	}

	return fmt.Sprintf("%v:%v:%v:%v",t.Hour, t.Minute, t.Second, t.Nano)
}
