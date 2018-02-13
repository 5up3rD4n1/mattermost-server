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
	Principal int64 `json:"principal"`
	UserName  string
	Password  string
	FullName  string
}


