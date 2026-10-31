package gateway

import (
	"sync"
)

type GatewayState int

const (
	StateDisconnected GatewayState = iota
	StateConnecting
	StateHandshaking
	StateReady
	StateResuming
	StateClosed
)

func (s GatewayState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateHandshaking:
		return "handshaking"
	case StateReady:
		return "ready"
	case StateResuming:
		return "resuming"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

type SessionState struct {
	mu               sync.RWMutex
	state            GatewayState
	sessionID        string
	resumeGatewayURL string
	sequence         int64
	hasSequence      bool
}

func NewSessionState() *SessionState {
	return &SessionState{
		state: StateDisconnected,
	}
}

func (s *SessionState) State() GatewayState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *SessionState) SetState(state GatewayState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

func (s *SessionState) SetSequence(seq int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sequence = seq
	s.hasSequence = true
}

func (s *SessionState) Sequence() (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sequence, s.hasSequence
}

func (s *SessionState) SetReady(sessionID, resumeGatewayURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionID = sessionID
	s.resumeGatewayURL = resumeGatewayURL
	s.state = StateReady
}

func (s *SessionState) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID
}

func (s *SessionState) ResumeGatewayURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resumeGatewayURL
}

func (s *SessionState) CanResume() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID != "" && s.hasSequence
}

func (s *SessionState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionID = ""
	s.resumeGatewayURL = ""
	s.sequence = 0
	s.hasSequence = false
	s.state = StateDisconnected
}
