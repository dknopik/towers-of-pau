package towersofpau

import (
	"crypto/rand"
	"errors"
	"math/big"

	blst "github.com/supranational/blst/bindings/go"
)

// UpdateTranscript adds our contribution to the ceremony
func UpdateTranscript(ceremony *Ceremony) error {
	for _, transcript := range ceremony.Transcripts {
		secret := createRandom()
		if err := UpdatePowersOfTau(transcript, secret); err != nil {
			return err
		}
		UpdateWitness(transcript, secret)
		// Clear secret
		b := make([]byte, 32)
		rand.Read(b)
		secret.SetBytes(b)
	}
	return nil
}

// SubgroupChecks verifies that a ceremony looks correctly
func SubgroupChecks(ceremony Ceremony) bool {
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

// UpdatePowersOfTau updates the powers of tau with a secret
func UpdatePowersOfTau(transcript *Transcript, secret *big.Int) error {
	sec := new(blst.Scalar).Deserialize(secret.Bytes())
	scalar := new(blst.Scalar).Deserialize(secret.Bytes())
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
func UpdateWitness(transcript *Transcript, secret *big.Int) {
	newProduct := transcript.Witness.RunningProducts[len(transcript.Witness.RunningProducts)-1]
	sec := new(blst.Scalar).Deserialize(secret.Bytes())
	newProduct = newProduct.Mult(sec)
	transcript.Witness.RunningProducts = append(transcript.Witness.RunningProducts, newProduct)
	newPk := new(blst.P2Affine).Deserialize(secret.Bytes())
	transcript.Witness.PotPubkeys = append(transcript.Witness.PotPubkeys, *newPk)
}

func createRandom() *big.Int {
	return big.NewInt(3)
}
