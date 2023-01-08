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

type Contribution struct {
	NumG1Powers  int
	NumG2Powers  int
	PowersOfTau  PowersOfTau
	BLSSignature *bls12381.PointG1
	PotPubkey    *bls12381.PointG2
}

func (t *Contribution) Copy() *Contribution {
	return &Contribution{
		NumG1Powers:  t.NumG1Powers,
		NumG2Powers:  t.NumG2Powers,
		PowersOfTau:  t.PowersOfTau.Copy(),
		BLSSignature: t.BLSSignature,
		PotPubkey:    t.PotPubkey,
	}
}

type JSONPowersOfTau struct {
	G1Powers []string
	G2Powers []string
}

type JSONContribution struct {
	NumG1Powers  int             `json:"numG1Powers"`
	NumG2Powers  int             `json:"numG2Powers"`
	PowersOfTau  JSONPowersOfTau `json:"powersOfTau"`
	BLSSignature string          `json:"blsSignature"`
	PotPubkey    string          `json:"potPubkey"`
}

func Deserialize(reader io.Reader) (*Contribution, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	jsoncontribution := JSONContribution{}
	err := decoder.Decode(&jsoncontribution)
	if err != nil {
		return nil, err
	}
	return DeserializeJSONContribution(jsoncontribution)
}

func DeserializePointG1(point string) (*bls12381.PointG1, error) {
	blsDec := make([]byte, 48)

	_, err := hex.Decode(blsDec, []byte(strings.TrimPrefix(point, "0x")))
	if err != nil {
		return nil, err
	}

	return bls12381.NewG1().FromCompressed(blsDec)
}

func DeserializePointG2(point string) (*bls12381.PointG2, error) {
	dec := make([]byte, 96)

	_, err := hex.Decode(dec, []byte(strings.TrimPrefix(point, "0x")))
	if err != nil {
		return nil, err
	}

	return bls12381.NewG2().FromCompressed(dec)
}

func DeserializeJSONContribution(jsoncontribution JSONContribution) (*Contribution, error) {
	contribution := Contribution{
		NumG1Powers: jsoncontribution.NumG1Powers,
		NumG2Powers: jsoncontribution.NumG2Powers,
		PowersOfTau: PowersOfTau{
			G1Powers: make([]*bls12381.PointG1, len(jsoncontribution.PowersOfTau.G1Powers)),
			G2Powers: make([]*bls12381.PointG2, len(jsoncontribution.PowersOfTau.G2Powers)),
		},
	}

	for i, power := range jsoncontribution.PowersOfTau.G1Powers {
		res, err := DeserializePointG1(power)
		if err != nil {
			return nil, err
		}

		contribution.PowersOfTau.G1Powers[i] = res
	}

	for i, power := range jsoncontribution.PowersOfTau.G2Powers {
		res, err := DeserializePointG2(power)
		if err != nil {
			return nil, err
		}

		contribution.PowersOfTau.G2Powers[i] = res
	}

	blsRes, err := DeserializePointG1(jsoncontribution.BLSSignature)
	if err != nil {
		return nil, err
	}

	contribution.BLSSignature = blsRes

	potRes, err := DeserializePointG2(jsoncontribution.PotPubkey)
	if err != nil {
		return nil, err
	}

	contribution.PotPubkey = potRes

	return &contribution, nil
}

func Serialize(writer io.Writer, contribution *Contribution) error {
	jsoncontribution, err := SerializeJSONContribution(contribution)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(jsoncontribution)
	if err != nil {
		return err
	}
	return nil
}

func SerializeJSONPointG1(point *bls12381.PointG1) string {
	return "0x" + hex.EncodeToString(G1.ToCompressed(point))
}

func SerializeJSONPointG2(point *bls12381.PointG2) string {
	return "0x" + hex.EncodeToString(G2.ToCompressed(point))
}

func SerializeJSONContribution(contribution *Contribution) (JSONContribution, error) {
	jsoncontribution := JSONContribution{
		NumG1Powers: contribution.NumG1Powers,
		NumG2Powers: contribution.NumG2Powers,
		PowersOfTau: JSONPowersOfTau{
			G1Powers: make([]string, len(contribution.PowersOfTau.G1Powers)),
			G2Powers: make([]string, len(contribution.PowersOfTau.G2Powers)),
		},
		BLSSignature: SerializeJSONPointG1(contribution.BLSSignature),
		PotPubkey:    SerializeJSONPointG2(contribution.PotPubkey),
	}

	for i, point := range contribution.PowersOfTau.G1Powers {
		jsoncontribution.PowersOfTau.G1Powers[i] = SerializeJSONPointG1(point)
	}

	for i, point := range contribution.PowersOfTau.G2Powers {
		jsoncontribution.PowersOfTau.G2Powers[i] = SerializeJSONPointG2(point)
	}

	return jsoncontribution, nil
}
