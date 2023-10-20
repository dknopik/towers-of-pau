package towersofpau

import (
	"encoding/hex"
	"encoding/json"

	"io"
	"strings"

	blst "github.com/supranational/blst/bindings/go"
)

type PowersOfTau struct {
	G1Powers []*blst.P1
	G2Powers []*blst.P2
}

func (p *PowersOfTau) Copy() PowersOfTau {
	p1 := make([]*blst.P1, 0, len(p.G1Powers))
	for _, pow := range p.G1Powers {
		p1 = append(p1, &(*pow))
	}
	p2 := make([]*blst.P2, 0, len(p.G2Powers))
	for _, pow := range p.G2Powers {
		p2 = append(p2, &(*pow))
	}
	return PowersOfTau{
		G1Powers: p1,
		G2Powers: p2,
	}
}

type Witness struct {
	RunningProducts []*blst.P1
	PotPubkeys      blst.P2Affines
}

func (w *Witness) Copy() *Witness {
	products := make([]*blst.P1, 0, len(w.RunningProducts))
	for _, p := range w.RunningProducts {
		products = append(products, &(*p))
	}
	return &Witness{
		RunningProducts: products,
		PotPubkeys:      w.PotPubkeys,
	}
}

type Transcript struct {
	NumG1Powers int
	NumG2Powers int
	PowersOfTau PowersOfTau
	Witness     *Witness
}

func (t *Transcript) Copy() *Transcript {
	return &Transcript{
		NumG1Powers: t.NumG1Powers,
		NumG2Powers: t.NumG2Powers,
		PowersOfTau: t.PowersOfTau.Copy(),
		Witness:     t.Witness.Copy(),
	}
}

type Ceremony struct {
	Transcripts                []*Transcript
	ParticipantIDs             []string
	ParticipantEcdsaSignatures []string
}

func (c *Ceremony) Copy() *Ceremony {
	transcripts := make([]*Transcript, 0, len(c.Transcripts))
	for _, t := range c.Transcripts {
		transcripts = append(transcripts, t.Copy())
	}
	return &Ceremony{
		Transcripts: transcripts,
	}
}

type JSONPowersOfTau struct {
	G1Powers []string
	G2Powers []string
}

type JSONWitness struct {
	RunningProducts []string `json:"runningProducts"`
	PotPubkeys      []string `json:"potPubkeys"`
	BLSSignatures   []string `json:"blsSignatures"`
}

type JSONTranscript struct {
	NumG1Powers int             `json:"numG1Powers"`
	NumG2Powers int             `json:"numG2Powers"`
	PowersOfTau JSONPowersOfTau `json:"powersOfTau"`
	Witness     JSONWitness     `json:"witness"`
}

type JSONCeremony struct {
	Transcripts                []JSONTranscript `json:"transcripts"`
	ParticipantIDs             []string         `json:"participantIds"`
	ParticipantEcdsaSignatures []string         `json:"participantEcdsaSignatures"`
}

func Deserialize(reader io.Reader) (*Ceremony, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var jsonceremony JSONCeremony
	if err := decoder.Decode(&jsonceremony); err != nil {
		return nil, err
	}
	return DeserializeJSONCeremony(jsonceremony)
}

func DeserializeTranscripts(jsontranscripts []JSONTranscript) ([]*Transcript, error) {
	var result []*Transcript
	for _, jsontranscript := range jsontranscripts {
		transcript := Transcript{
			NumG1Powers: jsontranscript.NumG1Powers,
			NumG2Powers: jsontranscript.NumG2Powers,
			PowersOfTau: PowersOfTau{
				G1Powers: make([]*blst.P1, len(jsontranscript.PowersOfTau.G1Powers)),
				G2Powers: make([]*blst.P2, len(jsontranscript.PowersOfTau.G2Powers)),
			},
			Witness: &Witness{
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

		result = append(result, &transcript)
	}
	return result, nil
}

func DeserializeJSONCeremony(jsonceremony JSONCeremony) (*Ceremony, error) {
	transcripts, err := DeserializeTranscripts(jsonceremony.Transcripts)
	if err != nil {
		return nil, err
	}
	return &Ceremony{
		transcripts,
		jsonceremony.ParticipantIDs,
		jsonceremony.ParticipantEcdsaSignatures,
	}, nil
}

func Serialize(writer io.Writer, ceremony *Ceremony) error {
	jsonceremony, err := SerializeJSONCeremony(ceremony)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(jsonceremony)
	if err != nil {
		return err
	}
	return nil
}

func SerializeTranscripts(transcripts []*Transcript) ([]JSONTranscript, error) {
	var result []JSONTranscript
	for _, transcript := range transcripts {
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

		for i, point := range transcript.PowersOfTau.G1Powers {
			jsontranscript.PowersOfTau.G1Powers[i] = "0x" + hex.EncodeToString(point.Compress())
		}

		for i, point := range transcript.PowersOfTau.G2Powers {
			jsontranscript.PowersOfTau.G2Powers[i] = "0x" + hex.EncodeToString(point.Compress())
		}

		for i, point := range transcript.Witness.RunningProducts {
			jsontranscript.Witness.RunningProducts[i] = "0x" + hex.EncodeToString(point.Compress())
		}

		for i, point := range transcript.Witness.PotPubkeys {
			jsontranscript.Witness.PotPubkeys[i] = "0x" + hex.EncodeToString(point.Compress())
		}

		result = append(result, jsontranscript)
	}
	return result, nil
}

func SerializeJSONCeremony(ceremony *Ceremony) (JSONCeremony, error) {
	transcripts, err := SerializeTranscripts(ceremony.Transcripts)
	return JSONCeremony{
		Transcripts:                transcripts,
		ParticipantIDs:             ceremony.ParticipantIDs,
		ParticipantEcdsaSignatures: ceremony.ParticipantEcdsaSignatures}, err
}
