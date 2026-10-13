package builder

import (
	"errors"
	"time"

	"github.com/sfossl/bfrm/types"
)

var (
	ErrEmbedTitleTooLong       = errors.New("bfrm/builder: embed title exceeds 256 characters")
	ErrEmbedDescriptionTooLong = errors.New("bfrm/builder: embed description exceeds 4096 characters")
	ErrEmbedTooManyFields      = errors.New("bfrm/builder: embed exceeds maximum of 25 fields")
	ErrEmbedFieldNameEmpty     = errors.New("bfrm/builder: embed field name cannot be empty")
	ErrEmbedFieldNameTooLong   = errors.New("bfrm/builder: embed field name exceeds 256 characters")
	ErrEmbedFieldValueEmpty    = errors.New("bfrm/builder: embed field value cannot be empty")
	ErrEmbedFieldValueTooLong  = errors.New("bfrm/builder: embed field value exceeds 1024 characters")
	ErrEmbedFooterTooLong      = errors.New("bfrm/builder: embed footer text exceeds 2048 characters")
	ErrEmbedAuthorTooLong      = errors.New("bfrm/builder: embed author name exceeds 256 characters")
	ErrEmbedTotalLengthExceeded = errors.New("bfrm/builder: embed total character count exceeds 6000")
)

type EmbedBuilder struct {
	embed types.Embed
	err   error
}

func NewEmbed() *EmbedBuilder {
	return &EmbedBuilder{
		embed: types.Embed{
			Type: "rich",
		},
	}
}

func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	if len(title) > 256 {
		b.err = ErrEmbedTitleTooLong
		return b
	}
	b.embed.Title = title
	return b
}

func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	if len(desc) > 4096 {
		b.err = ErrEmbedDescriptionTooLong
		return b
	}
	b.embed.Description = desc
	return b
}

func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

func (b *EmbedBuilder) SetColor(color int) *EmbedBuilder {
	b.embed.Color = color
	return b
}

func (b *EmbedBuilder) SetTimestamp(t time.Time) *EmbedBuilder {
	utc := t.UTC()
	b.embed.Timestamp = &utc
	return b
}

func (b *EmbedBuilder) SetFooter(text, iconURL string) *EmbedBuilder {
	if len(text) > 2048 {
		b.err = ErrEmbedFooterTooLong
		return b
	}
	b.embed.Footer = &types.EmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	b.embed.Image = &types.EmbedMedia{
		URL: url,
	}
	return b
}

func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &types.EmbedMedia{
		URL: url,
	}
	return b
}

func (b *EmbedBuilder) SetAuthor(name, url, iconURL string) *EmbedBuilder {
	if len(name) > 256 {
		b.err = ErrEmbedAuthorTooLong
		return b
	}
	b.embed.Author = &types.EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	if b.err != nil {
		return b
	}
	if len(b.embed.Fields) >= 25 {
		b.err = ErrEmbedTooManyFields
		return b
	}
	if name == "" {
		b.err = ErrEmbedFieldNameEmpty
		return b
	}
	if len(name) > 256 {
		b.err = ErrEmbedFieldNameTooLong
		return b
	}
	if value == "" {
		b.err = ErrEmbedFieldValueEmpty
		return b
	}
	if len(value) > 1024 {
		b.err = ErrEmbedFieldValueTooLong
		return b
	}
	b.embed.Fields = append(b.embed.Fields, types.EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

func (b *EmbedBuilder) Build() (types.Embed, error) {
	if b.err != nil {
		return types.Embed{}, b.err
	}

	total := len(b.embed.Title) + len(b.embed.Description)
	if b.embed.Footer != nil {
		total += len(b.embed.Footer.Text)
	}
	if b.embed.Author != nil {
		total += len(b.embed.Author.Name)
	}
	for _, f := range b.embed.Fields {
		total += len(f.Name) + len(f.Value)
	}

	if total > 6000 {
		return types.Embed{}, ErrEmbedTotalLengthExceeded
	}

	return b.embed, nil
}

