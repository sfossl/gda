package types

type GatewayOpcode int

const (
	OpcodeDispatch            GatewayOpcode = 0
	OpcodeHeartbeat           GatewayOpcode = 1
	OpcodeIdentify            GatewayOpcode = 2
	OpcodePresenceUpdate      GatewayOpcode = 3
	OpcodeVoiceStateUpdate    GatewayOpcode = 4
	OpcodeResume              GatewayOpcode = 6
	OpcodeReconnect           GatewayOpcode = 7
	OpcodeRequestGuildMembers GatewayOpcode = 8
	OpcodeInvalidSession      GatewayOpcode = 9
	OpcodeHello               GatewayOpcode = 10
	OpcodeHeartbeatACK        GatewayOpcode = 11
)

type VoiceOpcode int

const (
	VoiceOpcodeIdentify           VoiceOpcode = 0
	VoiceOpcodeSelectProtocol     VoiceOpcode = 1
	VoiceOpcodeReady              VoiceOpcode = 2
	VoiceOpcodeHeartbeat          VoiceOpcode = 3
	VoiceOpcodeSessionDescription VoiceOpcode = 4
	VoiceOpcodeSpeaking           VoiceOpcode = 5
	VoiceOpcodeHeartbeatACK       VoiceOpcode = 6
	VoiceOpcodeResume             VoiceOpcode = 7
	VoiceOpcodeHello              VoiceOpcode = 8
	VoiceOpcodeResumed            VoiceOpcode = 9
	VoiceOpcodeClientDisconnect   VoiceOpcode = 13
)

const (
	GatewayCloseUnknownError          = 4000
	GatewayCloseUnknownOpcode         = 4001
	GatewayCloseDecodeError           = 4002
	GatewayCloseNotAuthenticated      = 4003
	GatewayCloseAuthenticationFailed  = 4004
	GatewayCloseAlreadyAuthenticated  = 4005
	GatewayCloseInvalidSeq            = 4007
	GatewayCloseRateLimited           = 4008
	GatewayCloseSessionTimedOut       = 4009
	GatewayCloseInvalidShard          = 4010
	GatewayCloseShardingRequired      = 4011
	GatewayCloseInvalidAPIVersion     = 4012
	GatewayCloseInvalidIntents        = 4013
	GatewayCloseDisallowedIntents     = 4014
)

func IsReconnectable(closeCode int) bool {
	switch closeCode {
	case GatewayCloseAuthenticationFailed,
		GatewayCloseInvalidShard,
		GatewayCloseShardingRequired,
		GatewayCloseInvalidAPIVersion,
		GatewayCloseInvalidIntents,
		GatewayCloseDisallowedIntents:
		return false
	default:
		return true
	}
}

