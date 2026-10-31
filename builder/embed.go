package builder

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/sfossl/gda/types"
)

var (
	ErrEmbedTitleTooLong        = errors.New("embed title exceeds 256 characters")
	ErrEmbedDescriptionTooLong  = errors.New("embed description exceeds 4096 characters")
	ErrEmbedTooManyFields       = errors.New("embed exceeds maximum of 25 fields")
	ErrEmbedFieldNameEmpty      = errors.New("embed field name cannot be empty")
	ErrEmbedFieldNameTooLong    = errors.New("embed field name exceeds 256 characters")
	ErrEmbedFieldValueEmpty     = errors.New("embed field value cannot be empty")
	ErrEmbedFieldValueTooLong   = errors.New("embed field value exceeds 1024 characters")
	ErrEmbedFooterTooLong       = errors.New("embed footer text exceeds 2048 characters")
	ErrEmbedAuthorTooLong       = errors.New("embed author name exceeds 256 characters")
	ErrEmbedTotalLengthExceeded = errors.New("embed total character count exceeds 6000")
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
	if utf8.RuneCountInString(title) > 256 {
		b.err = ErrEmbedTitleTooLong
		return b
	}
	b.embed.Title = title
	return b
}

func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	if utf8.RuneCountInString(desc) > 4096 {
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

func (b *EmbedBuilder) SetFooter(text string, iconURL ...string) *EmbedBuilder {
	if utf8.RuneCountInString(text) > 2048 {
		b.err = ErrEmbedFooterTooLong
		return b
	}
	var icon string
	if len(iconURL) > 0 {
		icon = iconURL[0]
	}
	b.embed.Footer = &types.EmbedFooter{
		Text:    text,
		IconURL: icon,
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

func (b *EmbedBuilder) SetAuthor(name string, urls ...string) *EmbedBuilder {
	if utf8.RuneCountInString(name) > 256 {
		b.err = ErrEmbedAuthorTooLong
		return b
	}
	var url, iconURL string
	if len(urls) > 0 {
		url = urls[0]
	}
	if len(urls) > 1 {
		iconURL = urls[1]
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
	if utf8.RuneCountInString(name) > 256 {
		b.err = ErrEmbedFieldNameTooLong
		return b
	}
	if value == "" {
		b.err = ErrEmbedFieldValueEmpty
		return b
	}
	if utf8.RuneCountInString(value) > 1024 {
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

	total := utf8.RuneCountInString(b.embed.Title) + utf8.RuneCountInString(b.embed.Description)
	if b.embed.Footer != nil {
		total += utf8.RuneCountInString(b.embed.Footer.Text)
	}
	if b.embed.Author != nil {
		total += utf8.RuneCountInString(b.embed.Author.Name)
	}
	for _, f := range b.embed.Fields {
		total += utf8.RuneCountInString(f.Name) + utf8.RuneCountInString(f.Value)
	}

	if total > 6000 {
		return types.Embed{}, ErrEmbedTotalLengthExceeded
	}

	return b.embed, nil
}
