package types

import (
	"fmt"
	"strconv"
	"strings"
)

type AvatarOptions struct {
	Size    int
	Global  bool
	GuildID Snowflake
}

func parseAvatarArgs(args ...any) AvatarOptions {
	var opts AvatarOptions
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			opts.Size = v
		case bool:
			opts.Global = v
		case Snowflake:
			opts.GuildID = v
		case AvatarOptions:
			opts = v
		case *AvatarOptions:
			if v != nil {
				opts = *v
			}
		}
	}
	return opts
}

func (u *User) AvatarURL(args ...any) string {
	if u == nil {
		return ""
	}
	opts := parseAvatarArgs(args...)
	size := opts.Size
	if size <= 0 {
		size = 512
	}

	if u.Avatar != nil && *u.Avatar != "" {
		ext := "png"
		if strings.HasPrefix(*u.Avatar, "a_") {
			ext = "gif"
		}
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%d/%s.%s?size=%d", u.ID, *u.Avatar, ext, size)
	}

	var index int
	if u.Discriminator != "" && u.Discriminator != "0" {
		if d, err := strconv.Atoi(u.Discriminator); err == nil {
			index = d % 5
		}
	} else {
		index = int((uint64(u.ID) >> 22) % 6)
	}
	return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", index)
}

func (m *Member) AvatarURL(args ...any) string {
	if m == nil {
		return ""
	}
	opts := parseAvatarArgs(args...)
	size := opts.Size
	if size <= 0 {
		size = 512
	}

	guildID := m.GuildID
	if guildID == 0 && opts.GuildID != 0 {
		guildID = opts.GuildID
	}

	if !opts.Global && m.Avatar != nil && *m.Avatar != "" && guildID != 0 && m.User != nil && m.User.ID != 0 {
		ext := "png"
		if strings.HasPrefix(*m.Avatar, "a_") {
			ext = "gif"
		}
		return fmt.Sprintf("https://cdn.discordapp.com/guilds/%d/users/%d/avatars/%s.%s?size=%d", guildID, m.User.ID, *m.Avatar, ext, size)
	}

	if m.User != nil {
		return m.User.AvatarURL(opts)
	}
	return ""
}
