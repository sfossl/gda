package interaction

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"github.com/sfossl/gda/builder"
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrAlreadyResponded = errors.New("interaction already acknowledged")
	ErrModalAfterDefer  = errors.New("cannot send modal after deferring")
	ErrNoRestClient     = errors.New("rest client required for followups")
)

type Response struct {
	Content    string
	Embeds     []types.Embed
	Components []types.Component
	Ephemeral  bool
	TTS        bool
}

func splitCustomID(cid string) (string, string) {
	parts := strings.SplitN(cid, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func ephemeralFlag(ephemeral bool) types.MessageFlag {
	if ephemeral {
		return types.MessageFlagEphemeral
	}
	return 0
}

type Context struct {
	ctx          context.Context
	Interaction  *types.Interaction
	restClient   *rest.Client
	responded    atomic.Bool
	deferred     atomic.Bool
	responseChan chan types.InteractionResponse
	gatewayMode  bool
}

func NewContext(ctx context.Context, interaction *types.Interaction, restClient *rest.Client) *Context {
	return &Context{
		ctx:          ctx,
		Interaction:  interaction,
		restClient:   restClient,
		responseChan: make(chan types.InteractionResponse, 1),
	}
}

func NewGatewayContext(ctx context.Context, interaction *types.Interaction, restClient *rest.Client) *Context {
	return &Context{
		ctx:         ctx,
		Interaction: interaction,
		restClient:  restClient,
		gatewayMode: true,
	}
}

func (c *Context) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func (c *Context) ResponseChannel() <-chan types.InteractionResponse {
	return c.responseChan
}

func (c *Context) Defer(ephemeral bool) error {
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}
	c.deferred.Store(true)

	res := types.InteractionResponse{
		Type: types.CallbackTypeDeferredChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Flags: ephemeralFlag(ephemeral),
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- res:
			return nil
		default:
		}
	}

	if c.restClient != nil {
		return c.restClient.CreateInteractionResponse(c.Context(), c.Interaction.ID, c.Interaction.Token, res)
	}
	return ErrNoRestClient
}

func (c *Context) DeferUpdate() error {
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}
	c.deferred.Store(true)

	res := types.InteractionResponse{
		Type: types.CallbackTypeDeferredUpdateMessage,
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- res:
			return nil
		default:
		}
	}

	if c.restClient != nil {
		return c.restClient.CreateInteractionResponse(c.Context(), c.Interaction.ID, c.Interaction.Token, res)
	}
	return ErrNoRestClient
}

func (c *Context) Reply(res Response) error {
	flags := ephemeralFlag(res.Ephemeral)

	if c.deferred.Load() {
		if c.restClient == nil {
			return ErrNoRestClient
		}
		_, err := c.restClient.EditOriginalInteractionResponse(c.Context(), c.Interaction.AppID, c.Interaction.Token, rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
		})
		return err
	}

	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}

	ir := types.InteractionResponse{
		Type: types.CallbackTypeChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- ir:
			return nil
		default:
		}
	}

	if c.restClient != nil {
		return c.restClient.CreateInteractionResponse(c.Context(), c.Interaction.ID, c.Interaction.Token, ir)
	}
	return ErrNoRestClient
}

func (c *Context) UpdateOriginal(res Response) error {
	flags := ephemeralFlag(res.Ephemeral)

	if c.deferred.Load() {
		if c.restClient == nil {
			return ErrNoRestClient
		}
		_, err := c.restClient.EditOriginalInteractionResponse(c.Context(), c.Interaction.AppID, c.Interaction.Token, rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
		})
		return err
	}

	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}

	ir := types.InteractionResponse{
		Type: types.CallbackTypeUpdateMessage,
		Data: &types.InteractionResponseData{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- ir:
			return nil
		default:
		}
	}

	if c.restClient != nil {
		return c.restClient.CreateInteractionResponse(c.Context(), c.Interaction.ID, c.Interaction.Token, ir)
	}
	return ErrNoRestClient
}

func (c *Context) Modal(m builder.Modal) error {
	if c.deferred.Load() {
		return ErrModalAfterDefer
	}
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}

	ir := types.InteractionResponse{
		Type: types.CallbackTypeModal,
		Data: &types.InteractionResponseData{
			CustomID:   m.CustomID,
			Title:      m.Title,
			Components: m.Components,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- ir:
			return nil
		default:
		}
	}

	if c.restClient != nil {
		return c.restClient.CreateInteractionResponse(c.Context(), c.Interaction.ID, c.Interaction.Token, ir)
	}
	return ErrNoRestClient
}

func (c *Context) CommandName() string {
	if c.Interaction.Data != nil {
		return c.Interaction.Data.Name
	}
	return ""
}

func (c *Context) ComponentCustomID() string {
	if c.Interaction.Data != nil {
		return c.Interaction.Data.CustomID
	}
	return ""
}

func (c *Context) ComponentParam() string {
	_, param := splitCustomID(c.ComponentCustomID())
	return param
}

func (c *Context) User() *types.User {
	if c.Interaction.User != nil {
		return c.Interaction.User
	}
	if c.Interaction.Member != nil {
		return c.Interaction.Member.User
	}
	return nil
}

func (c *Context) Member() *types.Member {
	return c.Interaction.Member
}

func (c *Context) ChannelID() types.Snowflake {
	if c.Interaction.ChannelID != nil {
		return *c.Interaction.ChannelID
	}
	return 0
}

func (c *Context) GuildID() *types.Snowflake {
	return c.Interaction.GuildID
}

func (c *Context) Followup(res Response) (*types.Message, error) {
	if c.restClient == nil {
		return nil, ErrNoRestClient
	}
	return c.restClient.CreateFollowupMessage(c.Context(), c.Interaction.AppID, c.Interaction.Token, rest.MessagePayload{
		Content:    res.Content,
		Embeds:     res.Embeds,
		Components: res.Components,
		Flags:      ephemeralFlag(res.Ephemeral),
		TTS:        res.TTS,
	})
}

