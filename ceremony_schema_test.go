package towersofpau

import (
	"io"
	"os"
	"testing"
)

func TestInitialCeremony(t *testing.T) {
	file, err := os.Open("initialContribution4096.json")
	if err != nil {
		t.Error("unable to open")
	}
	contribution, err := Deserialize(file)
	if err != nil {
		t.Error("unable to decode", err.Error())
	}

	if contribution.NumG1Powers != len(contribution.PowersOfTau.G1Powers) {
		t.Error("wrong number of g1powers")
	}
	if contribution.NumG2Powers != len(contribution.PowersOfTau.G2Powers) {
		t.Error("wrong number of g2powers")
	}

	err = Serialize(io.Discard, contribution)
	if err != nil {
		t.Error("unable to serialize", err.Error())
	}
}
