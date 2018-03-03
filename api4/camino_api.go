package api4

import (
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
)

type CaminoRoutes struct {
	ApiRoot 	*mux.Router // 'api/v4'
	Contacts  	*mux.Router // 'api/v4/camino/contacts
}

func NewCaminoRoutes(root *mux.Router) *CaminoRoutes {
	routes := &CaminoRoutes{}

	routes.ApiRoot = root.PathPrefix(model.API_URL_CAMINO).Subrouter()
	routes.Contacts = routes.ApiRoot.PathPrefix("/contacts").Subrouter()

	return routes
}
