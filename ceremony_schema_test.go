package towersofpau

import (
	"encoding/json"
	"os"
	"testing"
)

func TestInitialCeremony(t *testing.T) {
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		println("unable to open")
		t.FailNow()
	}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	ceremony := Ceremony{
		[]Transcript{},
	}
	err = decoder.Decode(&ceremony)
	if err != nil {
		println("unable to decode", err.Error())
		t.FailNow()
	}
	if len(ceremony.Transcripts) == 0 {
		println("empty result")
		t.FailNow()
	}

	println("wat", len(ceremony.Transcripts))
}
