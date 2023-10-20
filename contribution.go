package towersofpau

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	blst "github.com/supranational/blst/bindings/go"
)

type BatchContribution struct {
	Contributions []*Contribution `json:"contributions"`
}

func (b *BatchContribution) Copy() *BatchContribution {
	contributions := make([]*Contribution, 0, len(b.Contributions))
	for _, t := range b.Contributions {
		contributions = append(contributions, t.Copy())
	}
	return &BatchContribution{
		Contributions: contributions,
	}
}

func (b *BatchContribution) SubgroupChecks() bool {
	for _, contribution := range b.Contributions {
		for _, p := range contribution.PowersOfTau.G1Powers {
			if !p.ToAffine().InG1() {
				return false
			}
		}
		for _, p := range contribution.PowersOfTau.G2Powers {
			if !p.ToAffine().InG2() {
				return false
			}
		}
		if !contribution.PotPubKey.ToAffine().InG2() {
			return false
		}
	}
	return true
}

type Contribution struct {
	NumG1Powers int
	NumG2Powers int
	PowersOfTau PowersOfTau
	PotPubKey   *blst.P2
}

func (t *Contribution) Copy() *Contribution {
	pk := *t.PotPubKey
	return &Contribution{
		NumG1Powers: t.NumG1Powers,
		NumG2Powers: t.NumG2Powers,
		PowersOfTau: t.PowersOfTau.Copy(),
		PotPubKey:   &pk,
	}
}

func (c *Contribution) UnmarshalJSON(data []byte) error {
	var v struct {
		NumG1Powers int             `json:"numG1Powers"`
		NumG2Powers int             `json:"numG2Powers"`
		PowersOfTau JSONPowersOfTau `json:"powersOfTau"`
		PotPubKey   string          `json:"potPubKey"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.NumG1Powers = v.NumG1Powers
	c.NumG2Powers = v.NumG2Powers
	c.PowersOfTau.G1Powers = make([]*blst.P1, len(v.PowersOfTau.G1Powers))
	c.PowersOfTau.G2Powers = make([]*blst.P2, len(v.PowersOfTau.G2Powers))

	bytes := make([][]byte, v.NumG1Powers)
	for i, power := range v.PowersOfTau.G1Powers {
		bytes[i] = make([]byte, 48)
		_, err := hex.Decode(bytes[i], []byte(strings.TrimPrefix(power, "0x")))
		if err != nil {
			return err
		}
	}
	resultP1 := new(blst.P1Affine).BatchUncompress(bytes)
	for i, affine := range resultP1 {
		c.PowersOfTau.G1Powers[i] = new(blst.P1)
		c.PowersOfTau.G1Powers[i].FromAffine(affine)
	}

	bytes = make([][]byte, v.NumG2Powers)
	for i, power := range v.PowersOfTau.G2Powers {
		bytes[i] = make([]byte, 96)
		_, err := hex.Decode(bytes[i], []byte(strings.TrimPrefix(power, "0x")))
		if err != nil {
			return err
		}
	}
	resultP2 := new(blst.P2Affine).BatchUncompress(bytes)
	for i, affine := range resultP2 {
		c.PowersOfTau.G2Powers[i] = new(blst.P2)
		c.PowersOfTau.G2Powers[i].FromAffine(affine)
	}

	pkBytes := make([]byte, 96)
	_, err := hex.Decode(pkBytes, []byte(strings.TrimPrefix(v.PotPubKey, "0x")))
	if err != nil {
		return err
	}
	potPubkey := new(blst.P2Affine).Uncompress(pkBytes)
	c.PotPubKey = new(blst.P2)
	c.PotPubKey.FromAffine(potPubkey)
	return nil
}

func (c *Contribution) MarshalJSON() ([]byte, error) {
	var v struct {
		NumG1Powers int             `json:"numG1Powers"`
		NumG2Powers int             `json:"numG2Powers"`
		PowersOfTau JSONPowersOfTau `json:"powersOfTau"`
		PotPubKey   string          `json:"potPubkey"`
	}

	v.NumG1Powers = len(c.PowersOfTau.G1Powers)
	v.NumG2Powers = len(c.PowersOfTau.G2Powers)
	v.PowersOfTau.G1Powers = make([]string, len(c.PowersOfTau.G1Powers))
	v.PowersOfTau.G2Powers = make([]string, len(c.PowersOfTau.G2Powers))

	for i, point := range c.PowersOfTau.G1Powers {
		v.PowersOfTau.G1Powers[i] = "0x" + hex.EncodeToString(point.Compress())
	}

	for i, point := range c.PowersOfTau.G2Powers {
		v.PowersOfTau.G2Powers[i] = "0x" + hex.EncodeToString(point.Compress())
	}

	v.PotPubKey = "0x" + hex.EncodeToString(c.PotPubKey.Compress())

	return json.Marshal(v)
}

// UpdateContribution adds our contribution to the ceremony
func UpdateContribution(contribution *BatchContribution) error {
	for _, transcript := range contribution.Contributions {
		rnd := createRandom()
		secret := common.LeftPadBytes(rnd.Bytes(), 32)
		if err := UpdatePowersOfTauFastContribution(transcript, secret); err != nil {
			return err
		}
		if err := UpdatePotPubkey(transcript, secret); err != nil {
			return err
		}
		// Clear secret
		rand.Read(secret)
	}
	return nil
}

func UpdatePowersOfTauFastContribution(transcript *Contribution, secret []byte) error {
	sec := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	one := make([]byte, 32)
	one[31] = 1
	scalar := new(blst.Scalar).Deserialize(one)
	if scalar == nil {
		return errors.New("failed to generate one scalar")
	}
	scalars := make([]*blst.Scalar, transcript.NumG1Powers)
	for i := 0; i < transcript.NumG1Powers; i++ {
		copy := *scalar
		scalars[i] = &copy
		var ok bool
		scalar, ok = scalar.Mul(sec)
		if !ok {
			panic("Scalar mult returned false")
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(transcript.NumG1Powers)
	for i := 0; i < transcript.NumG1Powers; i++ {
		go func(i int) {
			defer wg.Done()
			transcript.PowersOfTau.G1Powers[i] = transcript.PowersOfTau.G1Powers[i].Mult(scalars[i])
			if i < transcript.NumG2Powers {
				transcript.PowersOfTau.G2Powers[i] = transcript.PowersOfTau.G2Powers[i].Mult(scalars[i])
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func UpdatePotPubkey(transcript *Contribution, secret []byte) error {
	sec := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	newPk := new(blst.P2Affine).From(sec)
	if newPk == nil {
		return errors.New("invalid pk")
	}
	transcript.PotPubKey = new(blst.P2)
	transcript.PotPubKey.FromAffine(newPk)
	return nil
}
