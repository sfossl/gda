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
	raw := `{"id":"222333444555666777","channel_id":"333444555666777888","author":{"id":"444555666777888999","username":"author","discriminator":"0001"},"content":"hello world","timestamp":"2026-10-13T10:00:00Z","tts":false,"mention_everyone":false,"mentions":[],"embeds":[{"title":"Reminder","description":"Meeting in 5m","color":5793266}],"attachments":[],"components":[]}`
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
}
