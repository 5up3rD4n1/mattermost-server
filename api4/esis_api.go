package api4

import (
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
)

type EsisRoutes struct {
	ApiRoot 	*mux.Router // 'api/v4'
	Contacts  	*mux.Router // 'api/v4/esis/contacts
}

func NewEsisRoutes(root *mux.Router) *EsisRoutes {
	routes := &EsisRoutes{}

	routes.ApiRoot = root.PathPrefix(model.API_URL_ESIS).Subrouter()
	routes.Contacts = routes.ApiRoot.PathPrefix("/contacts").Subrouter()

	return routes
}
