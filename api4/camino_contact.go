package api4

import (
	"net/http"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/model/camino"
	l4g "github.com/alecthomas/log4go"
)

func (api *API) InitCaminoContacts() {
	api.CaminoRoutes.Contacts.Handle("", api.ApiSessionRequired(syncContact)).Methods("POST")
}

func syncContact(c *Context, w http.ResponseWriter, r *http.Request) {

	app := c.App

	if !app.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	contact := camino.ContactFromJson(r.Body)

	if err := app.SyncCaminoContact(contact, c.Session.UserId); err != nil {
		c.Err = err
		l4g.Error("Error trying to import camino contact", err)
		return
	}

	l4g.Debug("CAMINO Contact imported: Id = ", contact.Id)

	w.WriteHeader(http.StatusCreated)
	ReturnStatusOK(w)
}
