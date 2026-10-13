package types

type Intent uint64

const (
	IntentGuilds                      Intent = 1 << 0
	IntentGuildMembers                Intent = 1 << 1
	IntentGuildModeration             Intent = 1 << 2
	IntentGuildEmojisAndStickers      Intent = 1 << 3
	IntentGuildIntegrations           Intent = 1 << 4
	IntentGuildWebhooks               Intent = 1 << 5
	IntentGuildInvites                Intent = 1 << 6
	IntentGuildVoiceStates            Intent = 1 << 7
	IntentGuildPresences              Intent = 1 << 8
	IntentGuildMessages               Intent = 1 << 9
	IntentGuildMessageReactions       Intent = 1 << 10
	IntentGuildMessageTyping          Intent = 1 << 11
	IntentDirectMessages              Intent = 1 << 12
	IntentDirectMessageReactions      Intent = 1 << 13
	IntentDirectMessageTyping         Intent = 1 << 14
	IntentMessageContent              Intent = 1 << 15
	IntentGuildScheduledEvents        Intent = 1 << 16
	IntentAutoModerationConfiguration Intent = 1 << 20
	IntentAutoModerationExecution     Intent = 1 << 21
	IntentGuildMessagePolls           Intent = 1 << 24
	IntentDirectMessagePolls          Intent = 1 << 25

	IntentsNonPrivileged Intent = IntentGuilds |
		IntentGuildModeration |
		IntentGuildEmojisAndStickers |
		IntentGuildIntegrations |
		IntentGuildWebhooks |
		IntentGuildInvites |
		IntentGuildVoiceStates |
		IntentGuildMessages |
		IntentGuildMessageReactions |
		IntentGuildMessageTyping |
		IntentDirectMessages |
		IntentDirectMessageReactions |
		IntentDirectMessageTyping |
		IntentGuildScheduledEvents |
		IntentAutoModerationConfiguration |
		IntentAutoModerationExecution |
		IntentGuildMessagePolls |
		IntentDirectMessagePolls

	IntentsDefault Intent = IntentGuilds | IntentGuildMessages | IntentDirectMessages
)

func (i Intent) Has(target Intent) bool {
	return (i & target) == target
}

func (i Intent) Add(targets ...Intent) Intent {
	res := i
	for _, t := range targets {
		res |= t
	}
	return res
}

func (i Intent) Remove(targets ...Intent) Intent {
	res := i
	for _, t := range targets {
		res &^= t
	}
	return res
}

