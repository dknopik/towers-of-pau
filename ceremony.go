package towersofpau

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	blst "github.com/supranational/blst/bindings/go"
)

// UpdateTranscript adds our contribution to the ceremony
func UpdateTranscript(ceremony *Ceremony) error {
	for _, transcript := range ceremony.Transcripts {
		rnd := createRandom()
		secret := common.LeftPadBytes(rnd.Bytes(), 32)
		if err := UpdatePowersOfTau(transcript, secret); err != nil {
			return err
		}
		if err := UpdateWitness(transcript, secret); err != nil {
			return err
		}
		// Clear secret
		rand.Read(secret)
	}
	return nil
}

// UpdatePowersOfTau updates the powers of tau with a secret
func UpdatePowersOfTau(transcript *Transcript, secret []byte) error {
	sec := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	scalar := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	for i := 0; i < transcript.NumG1Powers; i++ {
		transcript.PowersOfTau.G1Powers[i] = transcript.PowersOfTau.G1Powers[i].Mult(scalar)
		if i < transcript.NumG2Powers {
			transcript.PowersOfTau.G2Powers[i] = transcript.PowersOfTau.G2Powers[i].Mult(scalar)
		}
		var ok bool
		scalar, ok = scalar.Mul(sec)
		if !ok {
			return errors.New("Scalar mult returned false")
		}
	}
	return nil
}

// UpdateWitness updates the witness with our secret.
func UpdateWitness(transcript *Transcript, secret []byte) error {
	newProduct := transcript.Witness.RunningProducts[len(transcript.Witness.RunningProducts)-1]
	sec := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	newProduct = newProduct.Mult(sec)
	transcript.Witness.RunningProducts = append(transcript.Witness.RunningProducts, newProduct)
	newPk := new(blst.P2Affine).From(sec)
	if newPk == nil {
		return errors.New("invalid pk")
	}
	transcript.Witness.PotPubkeys = append(transcript.Witness.PotPubkeys, *newPk)
	return nil
}

func createRandom() *big.Int {
	return big.NewInt(3)
}

// SubgroupChecks verifies that a ceremony looks correctly
func SubgroupChecksParticipant(ceremony Ceremony) bool {
	for _, transcript := range ceremony.Transcripts {
		for _, p := range transcript.PowersOfTau.G1Powers {
			if !p.ToAffine().InG1() {
				return false
			}
		}
		for _, p := range transcript.PowersOfTau.G2Powers {
			if !p.ToAffine().InG2() {
				return false
			}
		}
		for _, p := range transcript.Witness.RunningProducts {
			if !p.ToAffine().InG1() {
				return false
			}
		}
	}
	return true
}

// SubgroupChecksCoordinator verifies that a ceremony looks correctly
func SubgroupChecksCoordinator(ceremony *Ceremony) bool {
	for _, transcript := range ceremony.Transcripts {
		for _, p := range transcript.PowersOfTau.G1Powers {
			if !p.ToAffine().InG1() {
				return false
			}
		}
		for _, p := range transcript.PowersOfTau.G2Powers {
			if !p.ToAffine().InG2() {
				return false
			}
		}
		for _, p := range transcript.Witness.RunningProducts {
			if !p.ToAffine().InG1() {
				return false
			}
		}
		for _, p := range transcript.Witness.PotPubkeys {
			if !p.InG2() {
				return false
			}
		}
	}
	return true
}

// NonZeroCheck checks that no running_products are equal to infinity
func NonZeroCheck(ceremony *Ceremony) bool {
	for _, transcript := range ceremony.Transcripts {
		for _, p := range transcript.Witness.RunningProducts {
			_ = p
			// if !p.IsInfinite() { return false}
		}
	}
	return true
}

func PubkeyUniquenessCheck(ceremony *Ceremony) bool {
	keys := make(map[blst.P2Affine]struct{}, 0)
	var numKeys int
	for _, transcript := range ceremony.Transcripts {
		for _, key := range transcript.Witness.PotPubkeys {
			keys[key] = struct{}{}
			numKeys++
		}
	}
	return len(keys) == numKeys
}

func WitnessContinuityCheck(prevCeremony, newCeremony *Ceremony) bool {
	for index := range prevCeremony.Transcripts {
		oldWitness := prevCeremony.Transcripts[index].Witness
		newWitness := newCeremony.Transcripts[index].Witness
		// TODO check that we do a correct check
		if !p1ArrayEquals(oldWitness.RunningProducts, newWitness.RunningProducts) {
			return false
		}
		if !p2ArrayEquals(oldWitness.PotPubkeys, newWitness.PotPubkeys) {
			return false
		}
	}
	return true
}

func p1ArrayEquals(p1, p2 []*blst.P1) bool {
	if len(p1) != len(p2) {
		return false
	}
	for idx := range p1 {
		if !p1[idx].Equals(p2[idx]) {
			return false
		}
	}
	return true
}

func p2ArrayEquals(p1, p2 blst.P2Affines) bool {
	if len(p1) != len(p2) {
		return false
	}
	for idx := range p1 {
		if !p1[idx].Equals(&p2[idx]) {
			return false
		}
	}
	return true
}
