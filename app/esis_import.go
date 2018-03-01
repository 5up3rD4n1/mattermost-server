package app

import (
	"net/http"
	"strings"
	"regexp"
	"fmt"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/model/esis"
)

func (app *App) SyncEsisContact(contact *esis.Contact, userRequestorId string) *model.AppError {

	if contact.Type == esis.CONTACT_DIRECT {
		// create the users
		sender, sErr := app.createUserIfNotExists(contact.Sender)
		receiver, rErr := app.createUserIfNotExists(contact.Receiver)

		if sErr != nil { return sErr }
		if rErr != nil { return rErr }

		domainErr := validateUsersDomain(sender, receiver)

		if domainErr != nil { return domainErr }

		// create the teams with the user's domain
		team, tErr := app.createTeamFromUser(sender)

		if tErr != nil { return tErr }

		// join users to the corresponding team
		joinErr := app.joinEsisUsersToTeam(team, sender, receiver, userRequestorId)

		if joinErr != nil { return joinErr }

		channel, err := app.CreateDirectChannel(sender.Id, receiver.Id)
		if err != nil { return err }

		// By default the channel will not show up
		app.showDirectChannelToUser(channel.Id, sender.Id, receiver.Id)
		app.showDirectChannelToUser(channel.Id, receiver.Id, sender.Id)

	} else {
		detail := fmt.Sprintf("Principal id = %v", contact.Id)
		return model.NewAppError("App.EsisImport.createUserIfNotExists", "app.esis_import.group_mode_not_supported", nil, detail, http.StatusBadRequest)
	}

	return nil
}

func (app *App) createUserIfNotExists(principal *esis.Principal) (*model.User, *model.AppError) {
	if asUser := principal.AsUser; asUser != nil {
		if user, err := app.GetUserByEmail(principal.Handle); err != nil {
			newUser := &model.User{
				Email:          	principal.Handle,
				Username:       	asUser.Username,
				Password:       	asUser.Password,
				MessagingApiId: 	principal.Id,
				ReceiptWindowEnd: 	asUser.ReceiptWindowEnd.String(),
				ReceiptWindowStart: asUser.ReceiptWindowStart.String(),
			}
			return app.CreateUser(newUser)
		} else {
			return user, nil
		}
	}

	detail := fmt.Sprintf("Principal id = %v", principal.Id)
	return nil, model.NewAppError("App.EsisImport.createUserIfNotExists", "app.esis_import.group_mode_not_supported", nil, detail, http.StatusBadRequest)
}

func (app *App) createTeamFromUser(user *model.User) (*model.Team, *model.AppError) {
	senderDomain := strings.Split(user.Email, "@")[1]
	reg := regexp.MustCompile("[^A-Za-z0-9_-]+")
	cleanDomain := reg.ReplaceAllString(senderDomain, "")

	team, err := app.GetTeamByName(cleanDomain)

	if err != nil {
		newTeam := &model.Team{
			Name:  			cleanDomain,
			DisplayName:	cleanDomain,
			Email: 			"info@" + senderDomain,
			Type:  			model.TEAM_INVITE,
		}
		return app.CreateTeam(newTeam)
	}
	return team, nil
}

func (app *App) joinEsisUsersToTeam(team *model.Team, sender *model.User, receiver *model.User, userRequestorId string) *model.AppError {
	sErr := app.JoinUserToTeam(team, sender, userRequestorId)

	if sErr != nil { return sErr }

	rErr := app.JoinUserToTeam(team, receiver, userRequestorId)

	if rErr != nil {
		// Remove sender because the whole contact is not valid
		app.RemoveUserFromTeam(team.Id, sender.Id, userRequestorId)
		return rErr
	}
	return nil
}

func (app *App) showDirectChannelToUser(channelId, userId, otherUserId string) *model.AppError {
	preferences := model.Preferences{
		model.Preference{
			UserId:   userId,
			Category: model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW,
			Name:     channelId,
			Value:    "true",
		},
		model.Preference{
			UserId:   userId,
			Category: model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW,
			Name:     otherUserId,
			Value:    "true",
		},
	}

	return app.UpdatePreferences(userId, preferences)
}

func validateUsersDomain(sender *model.User, receiver *model.User) *model.AppError {
	senderDomain := strings.Split(sender.Email, "@")[1]
	receiverDomain := strings.Split(receiver.Email, "@")[1]

	if senderDomain != receiverDomain {
		return model.NewAppError("App.EsisImport.validateUsersDomain", "app.esis_import.cross_domain.not_allowed", nil, "", http.StatusBadRequest)
	}

	return nil
}
