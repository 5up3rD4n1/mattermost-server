package app

import (
	"net/http"
	"strings"
	"regexp"
	"fmt"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/model/camino"
	"github.com/alecthomas/log4go"
)

func (app *App) SyncCaminoContact(contact *camino.Contact, userRequestorId string) *model.AppError {

	if contact.Type == camino.CONTACT_DIRECT {
		return app.createDirectChannelFromContact(contact, userRequestorId)
	} else if contact.Type == camino.CONTACT_BROADCAST {
		return app.createGroupChannelFromContact(contact, userRequestorId)
	}

	detail := fmt.Sprintf("Principal id = %v", contact.Id)
	return model.NewAppError("App.CaminoImport.SyncCaminoContact", "app.camino_import.invalid_contact_type", nil, detail, http.StatusBadRequest)

}

func (app *App) createDirectChannelFromContact(contact *camino.Contact, userRequestorId string) *model.AppError {
	// create the users
	sender, sErr := app.createUserIfNotExists(contact.Sender.AsUser)
	receiver, rErr := app.createUserIfNotExists(contact.Receiver.AsUser)

	if sErr != nil { return sErr }
	if rErr != nil { return rErr }

	domainErr := validateUsersDomain(sender, receiver)

	if domainErr != nil { return domainErr }

	team, tErr := app.createTeamFromUser(sender)

	if tErr != nil { return tErr }

	users := []*model.User{sender, receiver}
	joinErr := app.joinCaminoUsersToTeam(users, team, userRequestorId)

	if joinErr != nil { return joinErr }

	channel, err := app.CreateDirectChannel(sender.Id, receiver.Id)
	if err != nil { return err }

	app.showDirectChannelToUser(channel.Id, sender.Id, receiver.Id)
	app.showDirectChannelToUser(channel.Id, receiver.Id, sender.Id)

	return nil
}

func (app *App) createGroupChannelFromContact(contact *camino.Contact, userRequestorId string) *model.AppError {


	caminoSender 	 := contact.Sender
	caminoReceiver := contact.Receiver

	if caminoSender.Type == camino.PRINCIPAL_USER && caminoReceiver.Type == camino.PRINCIPAL_GROUP {
		sender, sErr := app.createUserIfNotExists(caminoSender.AsUser)

		if sErr != nil { return sErr }

		team, tErr := app.createTeamFromUser(sender)

		if tErr != nil { return tErr }

		users, uErr := app.createUsersFromGroup(caminoReceiver.AsUserGroup)

		if uErr != nil { return uErr }

		jErr := app.joinCaminoUsersToTeam(append(users, sender), team, userRequestorId)

		if jErr != nil { return jErr }

		log4go.Info(caminoReceiver.AsUserGroup)
		groupName := caminoReceiver.AsUserGroup.Name
		channel, cErr := app.CreatePrivateChannel(groupName, team.Id, sender.Id)

		if cErr != nil { return cErr }

		app.addUsersToBroadcastChannel(users, channel, userRequestorId)
		app.AddUserToChannel(sender, channel)
	}
	return nil
}

func (app *App) createUsersFromGroup(group *camino.Group) ([]*model.User, *model.AppError) {
	var users []*model.User

	for _, usr := range group.Users {
		if newUser, err := app.createUserIfNotExists(usr); err == nil {
			users = append(users, newUser)
		} else {
			// TODO log soft error
		}
	}

	return users, nil
}

func (app *App) createUserIfNotExists(caminoUser *camino.User) (*model.User, *model.AppError) {
	log4go.Info(caminoUser)
	if usr, err := app.GetUserByEmail(caminoUser.Username); err != nil {
		newUser := &model.User{
			Email:          	caminoUser.Username,
			Username:       	caminoUser.Username,
			Password:       	caminoUser.Password,
			MessagingApiId: 	caminoUser.Id,
			ReceiptWindowEnd: 	caminoUser.ReceiptWindowEnd.String(),
			ReceiptWindowStart: caminoUser.ReceiptWindowStart.String(),
		}
		return app.CreateUser(newUser)
	} else {
		return usr, nil
	}
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

func (app *App) joinCaminoUsersToTeam(users []*model.User, team *model.Team, userRequestorId string) *model.AppError {
	var err *model.AppError
	for _, usr := range users {
		sErr := app.JoinUserToTeam(team, usr, userRequestorId)
		if sErr != nil { err = sErr }
	}

	return err
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

func (app *App) CreatePrivateChannel(name string, teamId string, creatorId string) (*model.Channel, *model.AppError) {

	channel := &model.Channel{
		TeamId: teamId,
		Name: strings.ToLower(name),
		DisplayName: name,
		CreatorId: creatorId,
		Type: model.CHANNEL_PRIVATE,
	}

	return app.CreateChannel(channel, true)
}

func (app *App) addUsersToBroadcastChannel(users []*model.User, channel *model.Channel, requestorId string) {
	usersMap := make(map[string]*model.User)

	for _, usr := range users {
		if _, err := app.AddChannelMember(usr.Id, channel, requestorId, ""); err == nil {
			usersMap[usr.Id] = usr
		}
	}

	for uId, usr := range usersMap {
		app.UpdateChannelMemberRoles(channel.Id, uId, model.CHANNEL_GUEST_ROLE_ID)
		app.AddUserToChannel(usr, channel)
	}
}

func validateUsersDomain(sender *model.User, receiver *model.User) *model.AppError {
	senderDomain := strings.Split(sender.Email, "@")[1]
	receiverDomain := strings.Split(receiver.Email, "@")[1]

	if senderDomain != receiverDomain {
		return model.NewAppError("App.CaminoImport.validateUsersDomain", "app.camino_import.cross_domain.not_allowed", nil, "", http.StatusBadRequest)
	}

	return nil
}

