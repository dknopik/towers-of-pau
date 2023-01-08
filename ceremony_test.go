package towersofpau

import (
	"os"
	"testing"
)

var contributionFiles = []string{
	"initialContribution4096.json",
	"initialContribution8192.json",
	"initialContribution16384.json",
	"initialContribution32768.json",
}

func TestContribution(t *testing.T) {
	for _, filename := range contributionFiles {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		contribution, err := Deserialize(file)
		if err != nil {
			t.Fatal(err)
		}

		secret := []byte("asdf")

		updatedContribution := contribution.Copy()

		if err := UpdateContribution(updatedContribution, secret); err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkContribution(t *testing.B) {
	for _, filename := range contributionFiles {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		contribution, err := Deserialize(file)
		if err != nil {
			t.Fatal(err)
		}
		t.ResetTimer()

		secret := []byte("asdf")

		if err := UpdateContribution(contribution, secret); err != nil {
			t.Fatal(err)
		}
	}
	panic("asdf")
}

func BenchmarkPairing(b *testing.B) {
	file, err := os.Open("initialContribution4096.json")
	if err != nil {
		b.Fatal(err)
	}
	contribution, err := Deserialize(file)
	if err != nil {
		b.Fatal(err)
	}
	g1Last := len(contribution.PowersOfTau.G1Powers) - 1
	if b.N < g1Last {
		g1Last = b.N
	}
	g2Last := len(contribution.PowersOfTau.G2Powers) - 1
	if b.N < g2Last {
		g2Last = b.N
	}
	pot := PowersOfTau{
		G1Powers: contribution.PowersOfTau.G1Powers[0:g1Last],
		G2Powers: contribution.PowersOfTau.G2Powers[0:g2Last],
	}
	_ = pot
}
