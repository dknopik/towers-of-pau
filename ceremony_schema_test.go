package towersofpau

import (
	"os"
	"testing"
)

func TestInitialCeremony(t *testing.T) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		println("unable to open")
		t.FailNow()
	}
	ceremony, err := Deserialize(file)
	if err != nil {
		println("unable to decode", err.Error())
		t.FailNow()
	}
	if len(ceremony.Transcripts) == 0 {
		println("empty result")
		t.FailNow()
	}
	if ceremony.Transcripts[0].NumG1Powers != len(ceremony.Transcripts[0].PowersOfTau.G1Powers) {
		println("wrong number of g1powers")
		t.FailNow()
	}
	if ceremony.Transcripts[0].NumG2Powers != len(ceremony.Transcripts[0].PowersOfTau.G2Powers) {
		println("wrong number of g2powers")
		t.FailNow()
	}
	err = Serialize(os.Stdout, ceremony)
	if err != nil {
		println("unable to serialize")
		t.FailNow()
	}
}
