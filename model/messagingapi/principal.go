package messagingapi

const (
	USER  = "USR"
	GROUP = "GRP"
)

type Principal struct {
	Id            int64  `json:"id"`
	Handle        string `json:"handle"`
	PrincipalType string `json:"type"`
	AsUser        *User
	AsGroup       *Group
}

type Group struct {
	Name      string
	Principal *Principal
	ApiUsers  *[]User
}

type User struct {
	Id		  			int64 	`json:"id"`
	UserName  			string 	`json:"userName"`
	Password  			string	`json:"password"`
	FullName  			string	`json:"fullName"`
	Principal 			int64 	`json:"principal"`
	ReceiptWindowStart 	string 	`json:"receiptWindowStart"`
	ReceiptWindowEnd 	string 	`json:"receiptWindowEnd"`
}
