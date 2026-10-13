package builder

import (
	"errors"
	"unicode/utf8"

	"github.com/sfossl/gda/types"
)

var (
	ErrActionRowEmpty              = errors.New("gda/builder: action row cannot be empty")
	ErrActionRowTooManyButtons     = errors.New("gda/builder: action row exceeds maximum of 5 buttons")
	ErrActionRowSelectNotAlone     = errors.New("gda/builder: select menu must be the only component in an action row")
	ErrActionRowTextInputNotAlone  = errors.New("gda/builder: text input must be the only component in an action row")
	ErrButtonCustomIDRequired      = errors.New("gda/builder: button custom_id is required for non-link buttons")
	ErrButtonCustomIDTooLong       = errors.New("gda/builder: button custom_id exceeds 100 characters")
	ErrButtonLabelTooLong          = errors.New("gda/builder: button label exceeds 80 characters")
	ErrButtonURLRequired           = errors.New("gda/builder: url is required for link button")
	ErrButtonURLNotAllowed         = errors.New("gda/builder: url is not allowed on non-link button")
	ErrSelectCustomIDRequired      = errors.New("gda/builder: select menu custom_id is required")
	ErrSelectCustomIDTooLong       = errors.New("gda/builder: select menu custom_id exceeds 100 characters")
	ErrSelectPlaceholderTooLong    = errors.New("gda/builder: select placeholder exceeds 150 characters")
	ErrSelectTooManyOptions        = errors.New("gda/builder: select menu exceeds maximum of 25 options")
	ErrSelectOptionLabelRequired   = errors.New("gda/builder: select option label is required")
	ErrSelectOptionValueRequired   = errors.New("gda/builder: select option value is required")
)

type ButtonStyle int

const (
	ButtonStylePrimary   ButtonStyle = 1
	ButtonStyleSecondary ButtonStyle = 2
	ButtonStyleSuccess   ButtonStyle = 3
	ButtonStyleDanger    ButtonStyle = 4
	ButtonStyleLink      ButtonStyle = 5
	ButtonStylePremium   ButtonStyle = 6
)

type ButtonBuilder struct {
	comp types.Component
	err  error
}

func NewButton() *ButtonBuilder {
	return &ButtonBuilder{
		comp: types.Component{
			Type:  types.ComponentTypeButton,
			Style: int(ButtonStylePrimary),
		},
	}
}

func (b *ButtonBuilder) SetStyle(style ButtonStyle) *ButtonBuilder {
	b.comp.Style = int(style)
	return b
}

func (b *ButtonBuilder) SetCustomID(customID string) *ButtonBuilder {
	if utf8.RuneCountInString(customID) > 100 {
		b.err = ErrButtonCustomIDTooLong
		return b
	}
	b.comp.CustomID = customID
	return b
}

func (b *ButtonBuilder) SetLabel(label string) *ButtonBuilder {
	if utf8.RuneCountInString(label) > 80 {
		b.err = ErrButtonLabelTooLong
		return b
	}
	b.comp.Label = label
	return b
}

func (b *ButtonBuilder) SetURL(url string) *ButtonBuilder {
	b.comp.URL = url
	return b
}

func (b *ButtonBuilder) SetDisabled(disabled bool) *ButtonBuilder {
	b.comp.Disabled = disabled
	return b
}

func (b *ButtonBuilder) SetEmoji(name string, id *types.Snowflake, animated bool) *ButtonBuilder {
	b.comp.Emoji = &types.ComponentEmoji{
		Name:     name,
		ID:       id,
		Animated: animated,
	}
	return b
}

func (b *ButtonBuilder) Build() (types.Component, error) {
	if b.err != nil {
		return types.Component{}, b.err
	}

	switch b.comp.Style {
	case int(ButtonStyleLink):
		if b.comp.URL == "" {
			return types.Component{}, ErrButtonURLRequired
		}
		if b.comp.CustomID != "" {
			return types.Component{}, ErrButtonURLNotAllowed
		}
	case int(ButtonStylePremium):
	default:
		if b.comp.CustomID == "" {
			return types.Component{}, ErrButtonCustomIDRequired
		}
		if b.comp.URL != "" {
			return types.Component{}, ErrButtonURLNotAllowed
		}
	}

	return b.comp, nil
}

type SelectMenuBuilder struct {
	comp types.Component
	err  error
}

func newSelectMenu(t types.ComponentType, customID string) *SelectMenuBuilder {
	b := &SelectMenuBuilder{
		comp: types.Component{
			Type:     t,
			CustomID: customID,
		},
	}
	if customID == "" {
		b.err = ErrSelectCustomIDRequired
	} else if utf8.RuneCountInString(customID) > 100 {
		b.err = ErrSelectCustomIDTooLong
	}
	return b
}

func NewStringSelect(customID string) *SelectMenuBuilder {
	return newSelectMenu(types.ComponentTypeStringSelect, customID)
}

func NewUserSelect(customID string) *SelectMenuBuilder {
	return newSelectMenu(types.ComponentTypeUserSelect, customID)
}

func NewRoleSelect(customID string) *SelectMenuBuilder {
	return newSelectMenu(types.ComponentTypeRoleSelect, customID)
}

func NewChannelSelect(customID string) *SelectMenuBuilder {
	return newSelectMenu(types.ComponentTypeChannelSelect, customID)
}

func (b *SelectMenuBuilder) SetPlaceholder(placeholder string) *SelectMenuBuilder {
	if utf8.RuneCountInString(placeholder) > 150 {
		b.err = ErrSelectPlaceholderTooLong
		return b
	}
	b.comp.Placeholder = placeholder
	return b
}

func (b *SelectMenuBuilder) SetValues(min, max int) *SelectMenuBuilder {
	b.comp.MinValues = &min
	b.comp.MaxValues = &max
	return b
}

func (b *SelectMenuBuilder) SetDisabled(disabled bool) *SelectMenuBuilder {
	b.comp.Disabled = disabled
	return b
}

func (b *SelectMenuBuilder) AddOption(label, value, description string, isDefault bool) *SelectMenuBuilder {
	if b.err != nil {
		return b
	}
	if len(b.comp.Options) >= 25 {
		b.err = ErrSelectTooManyOptions
		return b
	}
	if label == "" {
		b.err = ErrSelectOptionLabelRequired
		return b
	}
	if value == "" {
		b.err = ErrSelectOptionValueRequired
		return b
	}

	b.comp.Options = append(b.comp.Options, types.SelectOption{
		Label:       label,
		Value:       value,
		Description: description,
		Default:     isDefault,
	})
	return b
}

func (b *SelectMenuBuilder) Build() (types.Component, error) {
	if b.err != nil {
		return types.Component{}, b.err
	}
	return b.comp, nil
}

func NewActionRow(components ...types.Component) (types.Component, error) {
	if len(components) == 0 {
		return types.Component{}, ErrActionRowEmpty
	}

	hasSelect := false
	hasTextInput := false
	buttonCount := 0

	for _, c := range components {
		switch c.Type {
		case types.ComponentTypeButton:
			buttonCount++
		case types.ComponentTypeStringSelect,
			types.ComponentTypeUserSelect,
			types.ComponentTypeRoleSelect,
			types.ComponentTypeMentionableSelect,
			types.ComponentTypeChannelSelect:
			hasSelect = true
		case types.ComponentTypeTextInput:
			hasTextInput = true
		}
	}

	if hasSelect && len(components) > 1 {
		return types.Component{}, ErrActionRowSelectNotAlone
	}
	if hasTextInput && len(components) > 1 {
		return types.Component{}, ErrActionRowTextInputNotAlone
	}
	if buttonCount > 5 {
		return types.Component{}, ErrActionRowTooManyButtons
	}

	return types.Component{
		Type:       types.ComponentTypeActionRow,
		Components: components,
	}, nil
}

