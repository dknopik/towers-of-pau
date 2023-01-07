package towersofpau

import (
	"encoding/hex"
	"encoding/json"

	"io"
	"strings"

	bls12381 "github.com/kilic/bls12-381"
)

type PowersOfTau struct {
	G1Powers []*bls12381.PointG1
	G2Powers []*bls12381.PointG2
}

func (p *PowersOfTau) Copy() PowersOfTau {
	p1 := make([]*bls12381.PointG1, 0, len(p.G1Powers))
	p1 = append(p1, p.G1Powers...)
	p2 := make([]*bls12381.PointG2, 0, len(p.G2Powers))
	p2 = append(p2, p.G2Powers...)
	return PowersOfTau{
		G1Powers: p1,
		G2Powers: p2,
	}
}

type Witness struct {
	RunningProducts []*bls12381.PointG1
	PotPubkeys      []*bls12381.PointG2
}

func (w *Witness) Copy() *Witness {
	products := make([]*bls12381.PointG1, 0, len(w.RunningProducts))
	products = append(products, w.RunningProducts...)
	pubkeys := make([]*bls12381.PointG2, 0, len(w.PotPubkeys))
	pubkeys = append(pubkeys, w.PotPubkeys...)
	return &Witness{
		RunningProducts: products,
		PotPubkeys:      pubkeys,
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
	Transcripts []*Transcript
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
}

type JSONTranscript struct {
	NumG1Powers int             `json:"numG1Powers"`
	NumG2Powers int             `json:"numG2Powers"`
	PowersOfTau JSONPowersOfTau `json:"powersOfTau"`
	Witness     JSONWitness     `json:"witness"`
}

type JSONCeremony struct {
	Transcripts []JSONTranscript `json:"transcripts"`
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
	return DeserializeJSONCeremony(jsonceremony)
}

func DeserializeJSONCeremony(jsonceremony JSONCeremony) (*Ceremony, error) {
	ceremony := Ceremony{
		make([]*Transcript, 0, len(jsonceremony.Transcripts)),
	}
	for _, jsontranscript := range jsonceremony.Transcripts {
		transcript := Transcript{
			NumG1Powers: jsontranscript.NumG1Powers,
			NumG2Powers: jsontranscript.NumG2Powers,
			PowersOfTau: PowersOfTau{
				G1Powers: make([]*bls12381.PointG1, len(jsontranscript.PowersOfTau.G1Powers)),
				G2Powers: make([]*bls12381.PointG2, len(jsontranscript.PowersOfTau.G2Powers)),
			},
			Witness: &Witness{
				RunningProducts: make([]*bls12381.PointG1, len(jsontranscript.Witness.RunningProducts)),
				PotPubkeys:      make([]*bls12381.PointG2, len(jsontranscript.Witness.PotPubkeys)),
			},
		}

		for i, power := range jsontranscript.PowersOfTau.G1Powers {
			dec := make([]byte, 48)
			_, err := hex.Decode(dec, []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
			res, err := bls12381.NewG1().FromCompressed(dec)
			if err != nil {
				return nil, err
			}
			transcript.PowersOfTau.G1Powers[i] = res
		}

		for i, power := range jsontranscript.PowersOfTau.G2Powers {
			dec := make([]byte, 96)
			_, err := hex.Decode(dec, []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
			res, err := bls12381.NewG2().FromCompressed(dec)
			if err != nil {
				return nil, err
			}
			transcript.PowersOfTau.G2Powers[i] = res
		}

		for i, power := range jsontranscript.Witness.RunningProducts {
			dec := make([]byte, 48)
			_, err := hex.Decode(dec, []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
			res, err := bls12381.NewG1().FromCompressed(dec)
			if err != nil {
				return nil, err
			}
			transcript.Witness.RunningProducts[i] = res
		}

		for i, power := range jsontranscript.Witness.PotPubkeys {
			dec := make([]byte, 96)
			_, err := hex.Decode(dec, []byte(strings.TrimPrefix(power, "0x")))
			if err != nil {
				return nil, err
			}
			res, err := bls12381.NewG2().FromCompressed(dec)
			if err != nil {
				return nil, err
			}
			transcript.Witness.PotPubkeys[i] = res
		}

		ceremony.Transcripts = append(ceremony.Transcripts, &transcript)
	}
	return &ceremony, nil
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

func SerializeJSONCeremony(ceremony *Ceremony) (JSONCeremony, error) {
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

		for i, point := range transcript.PowersOfTau.G1Powers {
			jsontranscript.PowersOfTau.G1Powers[i] = "0x" + hex.EncodeToString(G1.ToCompressed(point))
		}

		for i, point := range transcript.PowersOfTau.G2Powers {
			jsontranscript.PowersOfTau.G2Powers[i] = "0x" + hex.EncodeToString(G2.ToCompressed(point))
		}

		for i, point := range transcript.Witness.RunningProducts {
			jsontranscript.Witness.RunningProducts[i] = "0x" + hex.EncodeToString(G1.ToCompressed(point))
		}

		for i, point := range transcript.Witness.PotPubkeys {
			jsontranscript.Witness.PotPubkeys[i] = "0x" + hex.EncodeToString(G2.ToCompressed(point))
		}

		jsonceremony.Transcripts = append(jsonceremony.Transcripts, jsontranscript)
	}
	return jsonceremony, nil
}
