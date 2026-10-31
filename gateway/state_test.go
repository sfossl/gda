package gateway

import (
	"testing"
)

func TestSessionState(t *testing.T) {
	s := NewSessionState()
	if s.State() != StateDisconnected {
		t.Fatalf("expected initial StateDisconnected, got %d", s.State())
	}
	if s.CanResume() {
		t.Fatalf("expected CanResume false initially")
	}

	s.SetState(StateConnecting)
	if s.State() != StateConnecting {
		t.Fatalf("expected StateConnecting, got %d", s.State())
	}

	s.SetSequence(42)
	seq, hasSeq := s.Sequence()
	if !hasSeq || seq != 42 {
		t.Fatalf("expected sequence 42, got %d", seq)
	}
	if s.CanResume() {
		t.Fatalf("expected CanResume false without sessionID")
	}

	s.SetReady("session_abc", "wss://gateway.discord.gg")
	if s.State() != StateReady {
		t.Fatalf("expected StateReady, got %d", s.State())
	}
	if s.SessionID() != "session_abc" {
		t.Fatalf("expected session_abc, got %s", s.SessionID())
	}
	if s.ResumeGatewayURL() != "wss://gateway.discord.gg" {
		t.Fatalf("expected resume URL, got %s", s.ResumeGatewayURL())
	}
	if !s.CanResume() {
		t.Fatalf("expected CanResume true")
	}

	s.Reset()
	if s.State() != StateDisconnected {
		t.Fatalf("expected StateDisconnected after reset")
	}
	if s.CanResume() {
		t.Fatalf("expected CanResume false after reset")
	}
	_, hasSeq = s.Sequence()
	if hasSeq {
		t.Fatalf("expected hasSequence false after reset")
	}
}
