package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	ErrCodeGeneralError                                 = 0
	ErrCodeUnknownAccount                               = 10001
	ErrCodeUnknownApplication                           = 10002
	ErrCodeUnknownChannel                               = 10003
	ErrCodeUnknownGuild                                 = 10004
	ErrCodeUnknownIntegration                           = 10005
	ErrCodeUnknownInvite                                = 10006
	ErrCodeUnknownMember                                = 10007
	ErrCodeUnknownMessage                               = 10008
	ErrCodeUnknownPermissionOverwrite                   = 10009
	ErrCodeUnknownProvider                              = 10010
	ErrCodeUnknownRole                                  = 10011
	ErrCodeUnknownToken                                 = 10012
	ErrCodeUnknownUser                                  = 10013
	ErrCodeUnknownEmoji                                 = 10014
	ErrCodeUnknownWebhook                               = 10015
	ErrCodeUnknownWebhookService                        = 10016
	ErrCodeUnknownSession                               = 10020
	ErrCodeUnknownBan                                   = 10026
	ErrCodeUnknownSKU                                   = 10027
	ErrCodeUnknownStoreListing                          = 10028
	ErrCodeUnknownEntitlement                           = 10029
	ErrCodeUnknownBuild                                 = 10030
	ErrCodeUnknownLobby                                 = 10031
	ErrCodeUnknownBranch                                = 10032
	ErrCodeUnknownStoreDirectoryLayout                  = 10033
	ErrCodeUnknownRedistributable                       = 10036
	ErrCodeUnknownGiftCode                              = 10038
	ErrCodeUnknownStream                                = 10049
	ErrCodeUnknownPremiumServerSubscribeCooldown        = 10050
	ErrCodeUnknownGuildTemplate                         = 10057
	ErrCodeUnknownDiscoverableServerCategory            = 10059
	ErrCodeUnknownSticker                               = 10060
	ErrCodeUnknownInteraction                           = 10062
	ErrCodeUnknownApplicationCommand                    = 10063
	ErrCodeUnknownVoiceState                            = 10065
	ErrCodeUnknownApplicationCommandPermissions         = 10066
	ErrCodeUnknownStageInstance                         = 10067
	ErrCodeUnknownGuildMemberVerificationForm           = 10068
	ErrCodeUnknownGuildWelcomeScreen                    = 10069
	ErrCodeUnknownGuildScheduledEvent                   = 10070
	ErrCodeUnknownGuildScheduledEventUser               = 10071
	ErrCodeUnknownTag                                   = 10087
	ErrCodeMissingAccess                                = 50001
	ErrCodeInvalidAccountType                           = 50002
	ErrCodeCannotExecuteOnDM                            = 50003
	ErrCodeEmbedDisabled                                = 50004
	ErrCodeCannotEditFromOtherUser                      = 50005
	ErrCodeCannotSendEmptyMessage                       = 50006
	ErrCodeCannotSendMessagesToUser                     = 50007
	ErrCodeCannotSendMessagesInVoiceChannel             = 50008
	ErrCodeChannelVerificationTooHigh                   = 50009
	ErrCodeOAuth2AppDoesNotHaveBot                      = 50010
	ErrCodeOAuth2AppLimitReached                        = 50011
	ErrCodeInvalidOAuth2State                           = 50012
	ErrCodeMissingPermissions                           = 50013
	ErrCodeInvalidToken                                 = 50014
	ErrCodeNoteTooLong                                  = 50015
	ErrCodeTooManyMessageDeletions                      = 50016
	ErrCodeCannotPinInOtherChannel                      = 50019
	ErrCodeInviteCodeInvalidOrTaken                     = 50020
	ErrCodeCannotExecuteOnSystemMessage                 = 50021
	ErrCodeCannotExecuteOnThisChannelType               = 50024
	ErrCodeInvalidOAuth2AccessToken                     = 50025
	ErrCodeMissingRequiredOAuth2Scope                   = 50026
	ErrCodeInvalidWebhookToken                          = 50027
	ErrCodeInvalidRole                                  = 50028
	ErrCodeInvalidRecipients                            = 50033
	ErrCodeOneOfTheMessagesTooOldToBulkDelete           = 50034
	ErrCodeInvalidFormBody                              = 50035
	ErrCodeInviteAcceptedToGuildBotNotIn                = 50036
	ErrCodeInvalidActivityAction                        = 50039
	ErrCodeInvalidAPIVersion                            = 50041
	ErrCodeFileUploadedExceedsMaximumSize               = 50045
	ErrCodeInvalidFileUploaded                          = 50046
	ErrCodeCannotSelfRedeemGift                         = 50054
	ErrCodePaymentSourceRequired                        = 50070
	ErrCodeCannotDeleteChannelRequiredForCommunity      = 50074
	ErrCodeCannotModifyStickerMessage                   = 50080
	ErrCodeInvalidStickerSent                           = 50081
	ErrCodeThreadArchivedOrLocked                       = 50083
	ErrCodeInvalidThreadNotificationSettings            = 50084
	ErrCodeBeforeEarlierThanThreadCreation              = 50085
	ErrCodeCommunityChannelsMustBeText                  = 50086
	ErrCodeEventEntityPastTime                          = 50091
	ErrCodeServerNotAvailableInLocation                 = 50095
	ErrCodeServerNeedsMonetizationEnabled               = 50097
	ErrCodeServerNeedsMoreBoosts                        = 50101
	ErrCodeRequestBodyContainsInvalidJSON               = 50109
	ErrCodeOwnershipCannotBeTransferredToBot            = 50132
	ErrCodeFailedToResizeAssetBelowLimit                = 50138
	ErrCodeCannotMixSubscriptionAndNonSubscriptionRoles = 50144
	ErrCodeCannotConvertBetweenPremiumAndNormalEmoji    = 50145
	ErrCodeUploadedFileNotValidImage                    = 50146
	ErrCodeVoiceMessagesMustHaveAudioOnly               = 50159
	ErrCodeVoiceMessagesMustBeSingleAudio               = 50160
	ErrCodeVoiceMessagesCannotHaveAdditionalContent     = 50161
	ErrCodeVoiceMessagesCannotHaveEmbeds                = 50162
	ErrCodeTwoFactorAuthRequired                        = 60003
	ErrCodeNoUsersWithDiscordTagExist                   = 80004
	ErrCodeReactionBlocked                              = 90001
	ErrCodeApplicationNotYetAvailable                   = 110001
	ErrCodeApplicationNameAlreadyExists                 = 110002
	ErrCodeApplicationInteractionTimedOut               = 10062
	ErrCodeStageAlreadyOpen                             = 150006
	ErrCodeCannotReplyWithoutReadHistory                = 160002
	ErrCodeThreadAlreadyCreatedForMessage               = 160004
	ErrCodeThreadLocked                                 = 160005
	ErrCodeMaxActiveThreadsReached                      = 160006
	ErrCodeMaxActiveAnnouncementThreadsReached          = 160007
	ErrCodeInvalidJSONForSticker                        = 170001
	ErrCodeInvalidAssetURL                              = 180000
	ErrCodeAutoModerationActionBlockedByRule            = 200000
	ErrCodeAutoModerationTitleBlockedByRule             = 200001
	ErrCodeAutoModerationWebhooksBlocked                = 200002
)

type APIError struct {
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Errors     json.RawMessage `json:"errors,omitempty"`
	HTTPStatus int             `json:"-"`
}

func (e *APIError) Error() string {
	if e.HTTPStatus > 0 {
		return fmt.Sprintf("discord rest: http %d: [%d] %s", e.HTTPStatus, e.Code, e.Message)
	}
	return fmt.Sprintf("discord rest: [%d] %s", e.Code, e.Message)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		return apiErr.HTTPStatus == http.StatusNotFound
	}
	return false
}

func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		return apiErr.HTTPStatus == http.StatusTooManyRequests
	}
	return false
}

func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		return apiErr.HTTPStatus == http.StatusUnauthorized
	}
	return false
}

func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		return apiErr.HTTPStatus == http.StatusForbidden
	}
	return false
}

func HasErrorCode(err error, code int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		return apiErr.Code == code
	}
	return false
}
