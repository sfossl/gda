package types

import (
	"encoding/json"
	"testing"
)

func TestGatewayPayloadUnmarshal(t *testing.T) {
	raw := `{"op":0,"d":{"v":10,"user":{"id":"123456789012345678","username":"testbot","discriminator":"0000","bot":true},"guilds":[],"session_id":"sess_123","resume_gateway_url":"wss://gateway.discord.gg"},"s":1,"t":"READY"}`
	var p GatewayPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if p.Op != OpcodeDispatch {
		t.Fatalf("expected opcode dispatch, got %d", p.Op)
	}
	if *p.T != EventReady {
		t.Fatalf("expected event READY, got %s", *p.T)
	}
	if *p.S != 1 {
		t.Fatalf("expected sequence 1, got %d", *p.S)
	}

	var ready ReadyEvent
	if err := json.Unmarshal(p.D, &ready); err != nil {
		t.Fatalf("unmarshal ready event error: %v", err)
	}
	if ready.SessionID != "sess_123" {
		t.Fatalf("expected session sess_123, got %s", ready.SessionID)
	}
	if ready.User.Username != "testbot" {
		t.Fatalf("expected username testbot, got %s", ready.User.Username)
	}
	if ready.User.ID != Snowflake(123456789012345678) {
		t.Fatalf("expected snowflake id, got %d", ready.User.ID)
	}
}

func TestInteractionUnmarshal(t *testing.T) {
	raw := `{"id":"111222333444555666","application_id":"999888777666555444","type":3,"data":{"custom_id":"snooze:15","component_type":2},"token":"interaction_token_123","version":1}`
	var in Interaction
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatalf("unmarshal interaction error: %v", err)
	}
	if in.Type != InteractionTypeMessageComponent {
		t.Fatalf("expected component interaction, got %d", in.Type)
	}
	if in.Data == nil || in.Data.CustomID != "snooze:15" {
		t.Fatalf("expected custom_id snooze:15")
	}
	if in.Token != "interaction_token_123" {
		t.Fatalf("expected token interaction_token_123, got %s", in.Token)
	}
}

func TestMessageUnmarshal(t *testing.T) {
	raw := `{"id":"222333444555666777","channel_id":"333444555666777888","author":{"id":"444555666777888999","username":"author","discriminator":"0001"},"content":"hello world","timestamp":"2026-10-13T10:00:00Z","tts":false,"mention_everyone":false,"mentions":[],"embeds":[{"title":"Reminder","description":"Meeting in 5m","color":5793266}],"attachments":[{"id":"555","filename":"clip.mp4","size":1024,"url":"https://cdn.discordapp.com/clip.mp4","proxy_url":"https://media.discordapp.net/clip.mp4","flags":4}],"components":[]}`
	var msg Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal message error: %v", err)
	}
	if msg.Content != "hello world" {
		t.Fatalf("expected content hello world, got %s", msg.Content)
	}
	if len(msg.Embeds) != 1 || msg.Embeds[0].Title != "Reminder" {
		t.Fatalf("expected 1 embed titled Reminder")
	}
	if len(msg.Attachments) != 1 || msg.Attachments[0].Flags == nil || *msg.Attachments[0].Flags != 4 {
		t.Fatalf("expected attachment with flag 4, got: %+v", msg.Attachments)
	}
}

func TestComponentsV2Unmarshal(t *testing.T) {
	raw := `{"type":17,"id":1,"accent_color":5793266,"spoiler":true,"components":[{"type":9,"id":2,"components":[{"type":10,"id":3,"content":"Section text"}],"accessory":{"type":11,"id":4,"media":{"url":"https://example.com/thumb.png"},"description":"thumb"}},{"type":14,"id":5,"divider":true,"spacing":1},{"type":12,"id":6,"items":[{"media":{"url":"https://example.com/gallery.png"},"description":"photo"}]},{"type":13,"id":7,"file":{"url":"attachment://doc.pdf"},"name":"doc.pdf","size":2048}]}`
	var comp Component
	if err := json.Unmarshal([]byte(raw), &comp); err != nil {
		t.Fatalf("unmarshal v2 component error: %v", err)
	}
	if comp.Type != ComponentTypeContainer {
		t.Fatalf("expected container, got %d", comp.Type)
	}
	if comp.ID == nil || *comp.ID != 1 {
		t.Fatalf("expected component id 1")
	}
	if comp.AccentColor == nil || *comp.AccentColor != 5793266 {
		t.Fatalf("expected accent color 5793266")
	}
	if comp.Spoiler == nil || !*comp.Spoiler {
		t.Fatalf("expected spoiler true")
	}
	if len(comp.Components) != 4 {
		t.Fatalf("expected 4 children, got %d", len(comp.Components))
	}
	sec := comp.Components[0]
	if sec.Type != ComponentTypeSection || sec.Accessory == nil {
		t.Fatalf("expected section with accessory")
	}
	if sec.Accessory.Type != ComponentTypeThumbnail || sec.Accessory.Media == nil || sec.Accessory.Media.URL != "https://example.com/thumb.png" {
		t.Fatalf("expected thumbnail accessory")
	}
	sep := comp.Components[1]
	if sep.Type != ComponentTypeSeparator || sep.Divider == nil || !*sep.Divider || sep.Spacing == nil || *sep.Spacing != int(SeparatorSpacingSmall) {
		t.Fatalf("expected separator with divider and spacing")
	}
	gal := comp.Components[2]
	if gal.Type != ComponentTypeMediaGallery || len(gal.Items) != 1 || gal.Items[0].Media.URL != "https://example.com/gallery.png" {
		t.Fatalf("expected media gallery item")
	}
	fil := comp.Components[3]
	if fil.Type != ComponentTypeFile || fil.File == nil || fil.Name != "doc.pdf" || fil.Size == nil || *fil.Size != 2048 {
		t.Fatalf("expected file component")
	}

	data, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("marshal v2 component error: %v", err)
	}
	var roundtrip Component
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("unmarshal roundtrip error: %v", err)
	}
	if roundtrip.Type != ComponentTypeContainer || len(roundtrip.Components) != 4 {
		t.Fatalf("roundtrip mismatch")
	}

	if MessageFlagIsComponentsV2 != 1<<15 {
		t.Fatalf("expected flag 1<<15, got %d", MessageFlagIsComponentsV2)
	}
}

func TestMessageReplyAndMentionsUnmarshal(t *testing.T) {
	raw := `{"id":"100","channel_id":"200","type":19,"content":"reply text","author":{"id":"300","username":"user","discriminator":"0001"},"message_reference":{"message_id":"99","channel_id":"200"},"referenced_message":{"id":"99","channel_id":"200","content":"original","author":{"id":"400","username":"parent","discriminator":"0001"}}}`
	var msg Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if msg.Type != 19 {
		t.Fatalf("expected message type 19, got %d", msg.Type)
	}
	if msg.MessageReference == nil || msg.MessageReference.MessageID == nil || *msg.MessageReference.MessageID != 99 {
		t.Fatalf("expected message_reference id 99")
	}
	if msg.ReferencedMessage == nil || msg.ReferencedMessage.Content != "original" {
		t.Fatalf("expected referenced message content original")
	}

	mentions := AllowedMentions{
		Parse: []string{"users"},
		Roles: []Snowflake{555},
	}
	mData, err := json.Marshal(mentions)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var unmarshaled AllowedMentions
	if err := json.Unmarshal(mData, &unmarshaled); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(unmarshaled.Parse) != 1 || len(unmarshaled.Roles) != 1 {
		t.Fatalf("allowed mentions roundtrip mismatch")
	}

	if EventGuildMemberAdd != "GUILD_MEMBER_ADD" || EventTypingStart != "TYPING_START" {
		t.Fatalf("expected event constants to be defined")
	}
}
