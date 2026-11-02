package types

import (
	"strings"
	"testing"
)

func TestAvatarURL_User(t *testing.T) {
	var nilUser *User
	if nilUser.AvatarURL() != "" {
		t.Fatalf("expected empty string for nil user")
	}

	hashStatic := "1234567890abcdef"
	uStatic := &User{
		ID:     123456789012345678,
		Avatar: &hashStatic,
	}
	url := uStatic.AvatarURL()
	expected := "https://cdn.discordapp.com/avatars/123456789012345678/1234567890abcdef.png?size=512"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}

	urlCustomSize := uStatic.AvatarURL(1024)
	expectedCustomSize := "https://cdn.discordapp.com/avatars/123456789012345678/1234567890abcdef.png?size=1024"
	if urlCustomSize != expectedCustomSize {
		t.Fatalf("expected %s, got %s", expectedCustomSize, urlCustomSize)
	}

	hashAnim := "a_abcdef1234567890"
	uAnim := &User{
		ID:     123456789012345678,
		Avatar: &hashAnim,
	}
	if !strings.HasSuffix(uAnim.AvatarURL(), ".gif?size=512") {
		t.Fatalf("expected .gif extension for animated avatar, got %s", uAnim.AvatarURL())
	}

	uDefaultNew := &User{
		ID:            80351110224678912,
		Discriminator: "0",
	}
	expectedIndex := int((uint64(80351110224678912) >> 22) % 6)
	if !strings.Contains(uDefaultNew.AvatarURL(), "/embed/avatars/") {
		t.Fatalf("expected default avatar fallback, got %s", uDefaultNew.AvatarURL())
	}
	_ = expectedIndex

	uDefaultLegacy := &User{
		ID:            80351110224678912,
		Discriminator: "1337",
	}
	if !strings.Contains(uDefaultLegacy.AvatarURL(), "/embed/avatars/2.png") {
		t.Fatalf("expected legacy index 2 (1337 %% 5), got %s", uDefaultLegacy.AvatarURL())
	}
}

func TestAvatarURL_Member(t *testing.T) {
	var nilMember *Member
	if nilMember.AvatarURL() != "" {
		t.Fatalf("expected empty string for nil member")
	}

	userHash := "userhash123"
	serverHash := "serverhash456"
	m := &Member{
		GuildID: 999,
		Avatar:  &serverHash,
		User: &User{
			ID:     111,
			Avatar: &userHash,
		},
	}

	if m.AvatarURL() != "https://cdn.discordapp.com/guilds/999/users/111/avatars/serverhash456.png?size=512" {
		t.Fatalf("unexpected member avatar URL: %s", m.AvatarURL())
	}

	if m.AvatarURL(true, 256) != "https://cdn.discordapp.com/avatars/111/userhash123.png?size=256" {
		t.Fatalf("unexpected forced global avatar URL: %s", m.AvatarURL(true, 256))
	}

	mNoServerAvatar := &Member{
		GuildID: 999,
		User: &User{
			ID:     111,
			Avatar: &userHash,
		},
	}
	if mNoServerAvatar.AvatarURL(false, 256) != "https://cdn.discordapp.com/avatars/111/userhash123.png?size=256" {
		t.Fatalf("expected fallback to global avatar when server avatar is missing: %s", mNoServerAvatar.AvatarURL(false, 256))
	}

	mNoAvatars := &Member{
		GuildID: 999,
		User: &User{
			ID:            80351110224678912,
			Discriminator: "0",
		},
	}
	if !strings.Contains(mNoAvatars.AvatarURL(false), "/embed/avatars/") {
		t.Fatalf("expected fallback to default discord avatar when both missing: %s", mNoAvatars.AvatarURL(false))
	}

	optsUrl := m.AvatarURL(AvatarOptions{Size: 1024, Global: true})
	if optsUrl != "https://cdn.discordapp.com/avatars/111/userhash123.png?size=1024" {
		t.Fatalf("unexpected AvatarOptions result: %s", optsUrl)
	}
}
