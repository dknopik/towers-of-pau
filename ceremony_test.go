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

	if !VerifyPairing(ceremony) {
		t.Fatal("Pairing check failed")
	}
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

	updatedCeremony := ceremony.Copy()
	if err := UpdateTranscript(updatedCeremony); err != nil {
		t.Fatal(err)
	}

	if !SubgroupChecksCoordinator(updatedCeremony) {
		t.Fatalf("Subgroup check failed")
	}

	if !NonZeroCheck(updatedCeremony) {
		t.Fatal("NonZero check failed")
	}

	if !WitnessContinuityCheck(ceremony, updatedCeremony) {
		t.Fatal("continuity check failed")
	}

	if !VerifyPairing(updatedCeremony) {
		t.Fatal("Pairing check failed")
	}
}
