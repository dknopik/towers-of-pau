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
	checkLength(ceremony, updatedCeremony)
}

func checkLength(prev, next *Ceremony) {
	for i, t := range prev.Transcripts {
		if len(t.Witness.PotPubkeys) != len(next.Transcripts[i].Witness.PotPubkeys)+1 {
			panic("invalid pot pubkeys")
		}

	}
}

func BenchmarkVerifyCeremonyPairing(t *testing.B) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		t.Fatal(err)
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		t.Fatal(err)
	}
	t.ResetTimer()
	if !VerifyPairing(ceremony) {
		t.Fatal("Pairing check failed")
	}
}

func BenchmarkContribution(t *testing.B) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		t.Fatal(err)
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		t.Fatal(err)
	}
	t.ResetTimer()
	if err := UpdateTranscript(ceremony); err != nil {
		t.Fatal(err)
	}
	panic("asdf")
}

func BenchmarkPairing(b *testing.B) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		b.Fatal(err)
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		b.Fatal(err)
	}
	g1Last := len(ceremony.Transcripts[0].PowersOfTau.G1Powers) - 1
	if b.N < g1Last {
		g1Last = b.N
	}
	g2Last := len(ceremony.Transcripts[0].PowersOfTau.G2Powers) - 1
	if b.N < g2Last {
		g2Last = b.N
	}
	pot := PowersOfTau{
		G1Powers: ceremony.Transcripts[0].PowersOfTau.G1Powers[0:g1Last],
		G2Powers: ceremony.Transcripts[0].PowersOfTau.G2Powers[0:g2Last],
	}
	verifyPairing(pot)
}
