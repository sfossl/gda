package types

import (
	"bytes"
	"strconv"
)

type Permission uint64

const (
	PermissionCreateInstantInvite               Permission = 1 << 0
	PermissionKickMembers                       Permission = 1 << 1
	PermissionBanMembers                        Permission = 1 << 2
	PermissionAdministrator                     Permission = 1 << 3
	PermissionManageChannels                    Permission = 1 << 4
	PermissionManageGuild                       Permission = 1 << 5
	PermissionAddReactions                      Permission = 1 << 6
	PermissionViewAuditLog                      Permission = 1 << 7
	PermissionPrioritySpeaker                   Permission = 1 << 8
	PermissionStream                            Permission = 1 << 9
	PermissionViewChannel                       Permission = 1 << 10
	PermissionSendMessages                      Permission = 1 << 11
	PermissionSendTTSMessages                   Permission = 1 << 12
	PermissionManageMessages                    Permission = 1 << 13
	PermissionEmbedLinks                        Permission = 1 << 14
	PermissionAttachFiles                       Permission = 1 << 15
	PermissionReadMessageHistory                Permission = 1 << 16
	PermissionMentionEveryone                   Permission = 1 << 17
	PermissionUseExternalEmojis                 Permission = 1 << 18
	PermissionViewGuildInsights                 Permission = 1 << 19
	PermissionConnect                           Permission = 1 << 20
	PermissionSpeak                             Permission = 1 << 21
	PermissionMuteMembers                       Permission = 1 << 22
	PermissionDeafenMembers                     Permission = 1 << 23
	PermissionMoveMembers                       Permission = 1 << 24
	PermissionUseVAD                            Permission = 1 << 25
	PermissionChangeNickname                    Permission = 1 << 26
	PermissionManageNicknames                   Permission = 1 << 27
	PermissionManageRoles                       Permission = 1 << 28
	PermissionManageWebhooks                    Permission = 1 << 29
	PermissionManageGuildExpressions            Permission = 1 << 30
	PermissionUseApplicationCommands            Permission = 1 << 31
	PermissionRequestToSpeak                    Permission = 1 << 32
	PermissionManageEvents                      Permission = 1 << 33
	PermissionManageThreads                     Permission = 1 << 34
	PermissionCreatePublicThreads               Permission = 1 << 35
	PermissionCreatePrivateThreads              Permission = 1 << 36
	PermissionUseExternalStickers               Permission = 1 << 37
	PermissionSendMessagesInThreads             Permission = 1 << 38
	PermissionUseEmbeddedActivities             Permission = 1 << 39
	PermissionModerateMembers                   Permission = 1 << 40
	PermissionViewCreatorMonetizationAnalytics  Permission = 1 << 41
	PermissionUseSoundboard                     Permission = 1 << 42
	PermissionCreateGuildExpressions            Permission = 1 << 43
	PermissionCreateEvents                      Permission = 1 << 44
	PermissionUseExternalSounds                 Permission = 1 << 45
	PermissionSendVoiceMessages                 Permission = 1 << 46
	PermissionSendPolls                         Permission = 1 << 49
)

func (p Permission) Has(target Permission) bool {
	if (p & PermissionAdministrator) == PermissionAdministrator {
		return true
	}
	return (p & target) == target
}

func (p Permission) Add(targets ...Permission) Permission {
	res := p
	for _, t := range targets {
		res |= t
	}
	return res
}

func (p Permission) Remove(targets ...Permission) Permission {
	res := p
	for _, t := range targets {
		res &^= t
	}
	return res
}

func (p Permission) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 22)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(p), 10)
	return append(buf, '"'), nil
}

func (p *Permission) UnmarshalJSON(b []byte) error {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*p = 0
		return nil
	}
	v, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	*p = Permission(v)
	return nil
}

