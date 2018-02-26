package api4

import (
	"net/http"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/model/messagingapi"
	l4g "github.com/alecthomas/log4go"
)

func (api *API) InitEsisContacts() {
	api.EsisRoutes.Contacts.Handle("", api.ApiSessionRequired(syncContact)).Methods("POST")
}

func syncContact(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_ADD_USER_TO_TEAM) {
		c.SetPermissionError(model.PERMISSION_ADD_USER_TO_TEAM)
		return
	}

	contact := messagingapi.ContactFromJson(r.Body)

	l4g.Debug(contact)

	// TODO: Process contact to create users, teams and channels
	// w.Write([]byte(contact.ToJson()))

	w.WriteHeader(http.StatusCreated)
	ReturnStatusOK(w)
}
