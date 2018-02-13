package app

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/model/messagingapi"
)

func (a *App) ValidatePrincipal(user *model.User) (*messagingapi.Principal, *model.AppError) {

	result := <-a.Srv.ApiStore.Principal().GetByHandle(&user.Email)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*messagingapi.Principal), nil
}
