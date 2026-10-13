package types

import "testing"

func TestIntentsBitwise(t *testing.T) {
	intents := IntentGuilds
	if !intents.Has(IntentGuilds) {
		t.Fatalf("expected to have IntentGuilds")
	}
	if intents.Has(IntentGuildMessages) {
		t.Fatalf("did not expect to have IntentGuildMessages")
	}

	intents = intents.Add(IntentGuildMessages, IntentDirectMessages)
	if !intents.Has(IntentGuildMessages) || !intents.Has(IntentDirectMessages) {
		t.Fatalf("expected to have added intents")
	}

	intents = intents.Remove(IntentGuildMessages)
	if intents.Has(IntentGuildMessages) {
		t.Fatalf("did not expect to have removed IntentGuildMessages")
	}
	if !intents.Has(IntentGuilds) || !intents.Has(IntentDirectMessages) {
		t.Fatalf("remaining intents should be preserved")
	}
}

func TestIntentsDefaultAndNonPrivileged(t *testing.T) {
	if !IntentsDefault.Has(IntentGuilds) {
		t.Fatalf("default intents missing IntentGuilds")
	}
	if IntentsNonPrivileged.Has(IntentMessageContent) {
		t.Fatalf("non-privileged intents should not contain IntentMessageContent")
	}
	if IntentsNonPrivileged.Has(IntentGuildMembers) {
		t.Fatalf("non-privileged intents should not contain IntentGuildMembers")
	}
	if IntentsNonPrivileged.Has(IntentGuildPresences) {
		t.Fatalf("non-privileged intents should not contain IntentGuildPresences")
	}
}

