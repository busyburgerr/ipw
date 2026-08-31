package profile

import (
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeStringsTrimsAndDedupes(t *testing.T) {
	got := normalizeStrings([]string{"Русский", "  Русский ", "English", "", "  ", "english"})
	want := []string{"Русский", "English"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDedupeUUIDsDropsNilAndRepeats(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	got := dedupeUUIDs([]uuid.UUID{a, b, a, uuid.Nil, b})
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("got %v, want [%s %s]", got, a, b)
	}
}

func TestAvailabilityValidation(t *testing.T) {
	valid := []Availability{AvailabilityAvailable, AvailabilityLimited, AvailabilityUnavailable, AvailabilityUnknown}
	for _, a := range valid {
		if !a.valid() {
			t.Errorf("%q should be valid", a)
		}
	}
	if Availability("busy").valid() {
		t.Error(`"busy" should be invalid`)
	}
}
