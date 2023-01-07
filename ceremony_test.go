package towersofpau

import (
	"os"
	"testing"
)

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
	_ = pot
}
