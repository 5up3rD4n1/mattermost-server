package messagingapi

import (
	"github.com/mattermost/mattermost-server/store"
	"net/http"
	"strings"
	"encoding/base64"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/model/messagingapi"
	"encoding/json"
	"io"
	"io/ioutil"
	"fmt"
)

type Store interface {
	Principal() PrincipalStore
}

type PrincipalStore interface {
	GetByHandle(userEmail *string) store.StoreChannel
}

type RequestHandler interface {
	MakeRequest(method, url string, params string) (*http.Response, *model.AppError)
}

type ApiStore struct {
	ServerUrl 	string
	Token 		string
	Client 	*http.Client
}

type PrincipalStoreImpl struct {
	ResourceURL string
	Handler 	RequestHandler
}

type ApiClient struct {
	Store 		*ApiStore
}

func NewApiStore(config *model.MessagingApiSettings, client *http.Client) Store {
	user := *config.Username
	password := *config.Password
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + password))

	return &ApiStore{
		config.ConnectionUrl + "/api",
		token,
		client,
	}
}


func (s *ApiStore) Principal() PrincipalStore {
	return &PrincipalStoreImpl{
		"/principals",
		&ApiClient{s},
	}
}


func (ps *PrincipalStoreImpl) GetByHandle(userEmail *string) store.StoreChannel {
	return store.Do(func (result *store.StoreResult){
		requestUrl := ps.ResourceURL + "/search"

		params :=  make(map[string]string)

		fmt.Println(*userEmail)

		params["handle"] = *userEmail

		response, err := ps.Handler.MakeRequest("POST", requestUrl, model.MapToJson(params))

		defer consumeAndClose(response)

		if err != nil {
			result.Err = err
			return
		}

		var principal messagingapi.Principal

		if dError := json.NewDecoder(response.Body).Decode(&principal); dError != nil {
			result.Err = model.NewAppError("MessagingApiRequest", "messaging.api.parse_error", nil, dError.Error(), http.StatusBadRequest)
		}
		result.Data = &principal
	})
}

func (api *ApiClient) MakeRequest(method, url string, params string) (*http.Response, *model.AppError) {

	fmt.Println("PARAMS", params)
	requestURL := api.Store.ServerUrl + url
	req, _ := http.NewRequest(method, requestURL, strings.NewReader(params))

	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json")
	//req.Header.Set("Authorization", "Basic " + api.Store.Token )

	resp, err := api.Store.Client.Do(req)

	if err != nil {
		return nil, model.NewAppError("MessagingApiRequest", "api.preference.save_preferences.decode.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	return resp, nil
}

// This is required to re-use the underlying connection and not take up file descriptors
func consumeAndClose(r *http.Response) {
	if r.Body != nil {
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}
}
