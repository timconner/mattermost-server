// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"strings"

	l4g "github.com/alecthomas/log4go"
	goi18n "github.com/nicksnyder/go-i18n/i18n"

	"github.com/mattermost/mattermost-server/model"
)

type InviteProvider struct {
}

const (
	CMD_INVITE   = "invite"
)

func init() {
	RegisterCommandProvider(&InviteProvider{})
}

func (me *InviteProvider) GetTrigger() string {
	return CMD_INVITE
}

func (me *InviteProvider) GetCommand(a *App, T goi18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CMD_INVITE,
		AutoComplete:     true,
		AutoCompleteDesc: T("Add a member to the channel"),
		AutoCompleteHint: T("@[username]"),
		DisplayName:      T("invite"),
	}
}

func (me *InviteProvider) DoCommand(a *App, args *model.CommandArgs, message string) *model.CommandResponse {
	channel, err := a.GetChannel(args.ChannelId)
	if err != nil {
		return &model.CommandResponse{Text: args.T("api.command_channel_rename.channel.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}

	if channel.Type == model.CHANNEL_PRIVATE && !a.SessionHasPermissionToTeam(args.Session, channel.TeamId, model.PERMISSION_MANAGE_TEAM) {
		return &model.CommandResponse{Text: args.T("You do not have permission to manage private channels."), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}
	
	if channel.Type == model.CHANNEL_GROUP || channel.Type == model.CHANNEL_DIRECT {
		return &model.CommandResponse{Text: args.T("You can't add someone to a direct message channel."), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}
	
	if len(message) == 0 {
		return &model.CommandResponse{Text: args.T("A message must be provided with the /invite command."), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}

	targetUsername := ""

	targetUsername = strings.SplitN(message, " ", 2)[0]
	targetUsername = strings.TrimPrefix(targetUsername, "@")

	var userProfile *model.User
	if result := <-a.Srv.Store.User().GetByUsername(targetUsername); result.Err != nil {
		l4g.Error(result.Err.Error())
		return &model.CommandResponse{Text: args.T("We couldn't find the user"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	} else {
		userProfile = result.Data.(*model.User)
	}

	_, err = a.AddChannelMember(userProfile.Id, channel, args.UserId, args.RootId)
	if err != nil {
		return &model.CommandResponse{Text: args.T(err.Id, map[string]interface{}{"Channel": model.DEFAULT_CHANNEL}), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}

	return &model.CommandResponse{}
}
