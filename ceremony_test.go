package towersofpau

import (
	"os"
	"testing"
)

func TestCeremonyChecks(t *testing.T) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		t.Fatal(err)
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		t.Fatal(err)
	}

	if !SubgroupChecksCoordinator(ceremony) {
		t.Fatalf("Subgroup check failed")
	}

	if !NonZeroCheck(ceremony) {
		t.Fatal("NonZero check failed")
	}

	/*
		// TODO enable when better initial ceremony is available
		if !PubkeyUniquenessCheck(ceremony) {
			t.Fatal("Pubkey uniqueness check failed")
		}
	*/
}

func TestParticipation(t *testing.T) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		t.Fatal(err)
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		t.Fatal(err)
	}

	if err := UpdateTranscript(ceremony); err != nil {
		t.Fatal(err)
	}

	if !SubgroupChecksCoordinator(ceremony) {
		t.Fatalf("Subgroup check failed")
	}

	if !NonZeroCheck(ceremony) {
		t.Fatal("NonZero check failed")
	}
}
