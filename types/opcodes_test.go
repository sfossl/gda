package types

import "testing"

func TestIsReconnectable(t *testing.T) {
	if !IsReconnectable(GatewayCloseUnknownError) {
		t.Fatalf("expected 4000 to be reconnectable")
	}
	if !IsReconnectable(GatewayCloseRateLimited) {
		t.Fatalf("expected 4008 to be reconnectable")
	}
	if IsReconnectable(GatewayCloseAuthenticationFailed) {
		t.Fatalf("expected 4004 not to be reconnectable")
	}
	if IsReconnectable(GatewayCloseInvalidIntents) {
		t.Fatalf("expected 4013 not to be reconnectable")
	}
	if IsReconnectable(GatewayCloseDisallowedIntents) {
		t.Fatalf("expected 4014 not to be reconnectable")
	}
}

