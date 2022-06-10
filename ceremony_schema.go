package towersofpau

import (
	"encoding/hex"
	"encoding/json"

	blst "github.com/supranational/blst/bindings/go"
	"io"
	"strings"
)

type PowersOfTau struct {
	G1Powers []*blst.P1
	G2Powers []*blst.P2
}

type Witness struct {
	RunningProducts []*blst.P1
	PotPubkeys      blst.P2Affines
}

type Transcript struct {
	NumG1Powers int
	NumG2Powers int
	PowersOfTau PowersOfTau
	Witness     Witness
}

type Ceremony struct {
	Transcripts []*Transcript
}

type JSONPowersOfTau struct {
	G1Powers []string
	G2Powers []string
}

type JSONWitness struct {
	RunningProducts []string
	PotPubkeys      []string
}

type JSONTranscript struct {
	NumG1Powers int
	NumG2Powers int
	PowersOfTau JSONPowersOfTau
	Witness     JSONWitness
}

type JSONCeremony struct {
	Transcripts []JSONTranscript
}

func Deserialize(reader io.Reader) (*Ceremony, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	jsonceremony := JSONCeremony{
		[]JSONTranscript{},
	}
	err := decoder.Decode(&jsonceremony)
	if err != nil {
		return nil, err
	}
	ceremony := Ceremony{
		make([]*Transcript, 0, len(jsonceremony.Transcripts)),
	}
	for _, jsontranscript := range jsonceremony.Transcripts {
		transcript := Transcript{
			NumG1Powers: jsontranscript.NumG1Powers,
			NumG2Powers: jsontranscript.NumG2Powers,
			PowersOfTau: PowersOfTau{
				G1Powers: make([]*blst.P1, len(jsontranscript.PowersOfTau.G1Powers)),
				G2Powers: make([]*blst.P2, len(jsontranscript.PowersOfTau.G2Powers)),
			},
			Witness: Witness{
				RunningProducts: make([]*blst.P1, len(jsontranscript.Witness.RunningProducts)),
				PotPubkeys:      make(blst.P2Affines, len(jsontranscript.Witness.PotPubkeys)),
			},
		}

		bytes := make([][]byte, jsontranscript.NumG1Powers)
		for i, power := range jsontranscript.PowersOfTau.G1Powers {
			bytes[i] = make([]byte, 48)
			_, err := hex.Decode(bytes[i], []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
		}
		resultP1 := new(blst.P1Affine).BatchUncompress(bytes)
		for i, affine := range resultP1 {
			transcript.PowersOfTau.G1Powers[i] = new(blst.P1)
			transcript.PowersOfTau.G1Powers[i].FromAffine(affine)
		}

		bytes = make([][]byte, jsontranscript.NumG2Powers)
		for i, power := range jsontranscript.PowersOfTau.G2Powers {
			bytes[i] = make([]byte, 96)
			_, err := hex.Decode(bytes[i], []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
		}
		resultP2 := new(blst.P2Affine).BatchUncompress(bytes)
		for i, affine := range resultP2 {
			transcript.PowersOfTau.G2Powers[i] = new(blst.P2)
			transcript.PowersOfTau.G2Powers[i].FromAffine(affine)
		}

		bytes = make([][]byte, len(jsontranscript.Witness.RunningProducts))
		for i, power := range jsontranscript.Witness.RunningProducts {
			bytes[i] = make([]byte, 48)
			_, err := hex.Decode(bytes[i], []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
		}
		resultP1 = new(blst.P1Affine).BatchUncompress(bytes)
		for i, affine := range resultP1 {
			transcript.Witness.RunningProducts[i] = new(blst.P1)
			transcript.Witness.RunningProducts[i].FromAffine(affine)
		}

		bytes = make([][]byte, len(jsontranscript.Witness.PotPubkeys))
		for i, power := range jsontranscript.Witness.PotPubkeys {
			bytes[i] = make([]byte, 96)
			_, err := hex.Decode(bytes[i], []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
		}
		resultP2 = new(blst.P2Affine).BatchUncompress(bytes)
		for i, affine := range resultP2 {
			transcript.Witness.PotPubkeys[i] = *affine
		}

		ceremony.Transcripts = append(ceremony.Transcripts, &transcript)
	}
	return &ceremony, nil
}

func Serialize(writer io.Writer, ceremony *Ceremony) error {
	jsonceremony := JSONCeremony{
		make([]JSONTranscript, 0, len(ceremony.Transcripts)),
	}
	for _, transcript := range ceremony.Transcripts {
		jsontranscript := JSONTranscript{
			NumG1Powers: transcript.NumG1Powers,
			NumG2Powers: transcript.NumG2Powers,
			PowersOfTau: JSONPowersOfTau{
				G1Powers: make([]string, len(transcript.PowersOfTau.G1Powers)),
				G2Powers: make([]string, len(transcript.PowersOfTau.G2Powers)),
			},
			Witness: JSONWitness{
				RunningProducts: make([]string, len(transcript.Witness.RunningProducts)),
				PotPubkeys:      make([]string, len(transcript.Witness.PotPubkeys)),
			},
		}

		jsonceremony.Transcripts = append(jsonceremony.Transcripts, jsontranscript)
	}
	encoder := json.NewEncoder(writer)
	err := encoder.Encode(jsonceremony)
	if err != nil {
		return err
	}
	return nil
}
