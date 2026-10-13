package builder

import (
	"errors"

	"github.com/sfossl/bfrm/types"
)

var (
	ErrModalCustomIDRequired       = errors.New("bfrm/builder: modal custom_id is required")
	ErrModalCustomIDTooLong        = errors.New("bfrm/builder: modal custom_id exceeds 100 characters")
	ErrModalTitleRequired          = errors.New("bfrm/builder: modal title is required")
	ErrModalTitleTooLong           = errors.New("bfrm/builder: modal title exceeds 45 characters")
	ErrModalNoComponents          = errors.New("bfrm/builder: modal must contain at least one text input")
	ErrModalTooManyComponents      = errors.New("bfrm/builder: modal exceeds maximum of 5 text inputs")
	ErrTextInputCustomIDRequired   = errors.New("bfrm/builder: text input custom_id is required")
	ErrTextInputCustomIDTooLong    = errors.New("bfrm/builder: text input custom_id exceeds 100 characters")
	ErrTextInputLabelRequired      = errors.New("bfrm/builder: text input label is required")
	ErrTextInputLabelTooLong       = errors.New("bfrm/builder: text input label exceeds 45 characters")
	ErrTextInputPlaceholderTooLong = errors.New("bfrm/builder: text input placeholder exceeds 100 characters")
	ErrTextInputValueTooLong       = errors.New("bfrm/builder: text input value exceeds 4000 characters")
)

type TextInputStyle int

const (
	TextInputStyleShort     TextInputStyle = 1
	TextInputStyleParagraph TextInputStyle = 2
)

type TextInputBuilder struct {
	comp types.Component
	err  error
}

func NewTextInput(customID, label string, style TextInputStyle) *TextInputBuilder {
	b := &TextInputBuilder{
		comp: types.Component{
			Type:     types.ComponentTypeTextInput,
			CustomID: customID,
			Label:    label,
			Style:    int(style),
		},
	}
	if customID == "" {
		b.err = ErrTextInputCustomIDRequired
	} else if len(customID) > 100 {
		b.err = ErrTextInputCustomIDTooLong
	}
	if label == "" {
		b.err = ErrTextInputLabelRequired
	} else if len(label) > 45 {
		b.err = ErrTextInputLabelTooLong
	}
	return b
}

func (b *TextInputBuilder) SetPlaceholder(placeholder string) *TextInputBuilder {
	if len(placeholder) > 100 {
		b.err = ErrTextInputPlaceholderTooLong
		return b
	}
	b.comp.Placeholder = placeholder
	return b
}

func (b *TextInputBuilder) SetValue(value string) *TextInputBuilder {
	if len(value) > 4000 {
		b.err = ErrTextInputValueTooLong
		return b
	}
	b.comp.Label = value
	return b
}

func (b *TextInputBuilder) SetLength(min, max int) *TextInputBuilder {
	b.comp.MinValues = &min
	b.comp.MaxValues = &max
	return b
}

func (b *TextInputBuilder) Build() (types.Component, error) {
	if b.err != nil {
		return types.Component{}, b.err
	}
	return b.comp, nil
}

type ModalBuilder struct {
	customID   string
	title      string
	components []types.Component
	err        error
}

func NewModal(customID, title string) *ModalBuilder {
	b := &ModalBuilder{
		customID: customID,
		title:    title,
	}
	if customID == "" {
		b.err = ErrModalCustomIDRequired
	} else if len(customID) > 100 {
		b.err = ErrModalCustomIDTooLong
	}
	if title == "" {
		b.err = ErrModalTitleRequired
	} else if len(title) > 45 {
		b.err = ErrModalTitleTooLong
	}
	return b
}

func (b *ModalBuilder) AddTextInput(input *TextInputBuilder) *ModalBuilder {
	if b.err != nil {
		return b
	}
	if len(b.components) >= 5 {
		b.err = ErrModalTooManyComponents
		return b
	}
	comp, err := input.Build()
	if err != nil {
		b.err = err
		return b
	}
	row, err := NewActionRow(comp)
	if err != nil {
		b.err = err
		return b
	}
	b.components = append(b.components, row)
	return b
}

type Modal struct {
	CustomID   string            `json:"custom_id"`
	Title      string            `json:"title"`
	Components []types.Component `json:"components"`
}

func (b *ModalBuilder) Build() (Modal, error) {
	if b.err != nil {
		return Modal{}, b.err
	}
	if len(b.components) == 0 {
		return Modal{}, ErrModalNoComponents
	}
	return Modal{
		CustomID:   b.customID,
		Title:      b.title,
		Components: b.components,
	}, nil
}

