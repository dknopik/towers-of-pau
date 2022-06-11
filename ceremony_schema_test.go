package towersofpau

import (
	"io"
	"os"
	"testing"
)

func TestInitialCeremony(t *testing.T) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		t.Error("unable to open")
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		t.Error("unable to decode", err.Error())
	}
	if len(ceremony.Transcripts) == 0 {
		t.Error("empty result")
	}
	if ceremony.Transcripts[0].NumG1Powers != len(ceremony.Transcripts[0].PowersOfTau.G1Powers) {
		t.Error("wrong number of g1powers")
	}
	if ceremony.Transcripts[0].NumG2Powers != len(ceremony.Transcripts[0].PowersOfTau.G2Powers) {
		t.Error("wrong number of g2powers")
	}

	err = Serialize(io.Discard, ceremony)
	if err != nil {
		t.Error("unable to serialize", err.Error())
	}
}
