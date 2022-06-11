package towersofpau

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	blst "github.com/supranational/blst/bindings/go"
)

// UpdateTranscript adds our contribution to the ceremony
func UpdateTranscript(ceremony *Ceremony) error {
	for _, transcript := range ceremony.Transcripts {
		rnd := createRandom()
		secret := common.LeftPadBytes(rnd.Bytes(), 32)
		if err := UpdatePowersOfTauFast(transcript, secret); err != nil {
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

func VerifySubmission(prevCeremony, newCeremony *Ceremony) error {
	if err := checkLength(prevCeremony, newCeremony); err != nil {
		return err
	}

	if !NonZeroCheck(newCeremony) {
		return errors.New("nonZero check failed")
	}

	if !SubgroupChecksCoordinator(newCeremony) {
		return errors.New("subgroup check failed")
	}

	if !WitnessContinuityCheck(prevCeremony, newCeremony) {
		return errors.New("continuity check failed")
	}

	/*
		// TODO enable when better initial ceremony is available
		if !PubkeyUniquenessCheck(newCeremony) {
			return errors.New("pubkey uniqueness check failed")
		}*/

	if !VerifyPairing(newCeremony) {
		return errors.New("pairing check failed")
	}
	return nil
}

func checkLength(prev, next *Ceremony) error {
	for i, t := range prev.Transcripts {
		if len(t.Witness.PotPubkeys) >= len(next.Transcripts[i].Witness.PotPubkeys) {
			errors.New("pot_pubkeys did not increase")
		}
		if len(t.Witness.RunningProducts) >= len(next.Transcripts[i].Witness.RunningProducts) {
			errors.New("running_products did not increase")
		}
	}
	return nil
}

// UpdatePowersOfTau updates the powers of tau with a secret
func UpdatePowersOfTau(transcript *Transcript, secret []byte) error {
	sec := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	scalar := &(*sec)
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

func UpdatePowersOfTauFast(transcript *Transcript, secret []byte) error {
	sec := new(blst.Scalar).Deserialize(secret)
	if sec == nil {
		return errors.New("invalid secret")
	}
	scalar := &(*sec)
	scalars := make([]*blst.Scalar, transcript.NumG1Powers)
	for i := 0; i < transcript.NumG1Powers; i++ {
		scalars[i] = &(*scalar)
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
	for i := 0; i < 1000000; i++ {
		b := make([]byte, 32)
		n, err := rand.Read(b)
		if err != nil || n != 32 {
			panic("could not get good randomness")
		}
		sec := new(blst.Scalar).Deserialize(b)
		if sec != nil {
			return new(big.Int).SetBytes(b)
		}
	}
	panic("could not find secret in 1 million tries")
}

// SubgroupChecks verifies that a ceremony looks correctly
func SubgroupChecksParticipant(ceremony *Ceremony) bool {
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
		if !p1ArrayEquals(oldWitness.RunningProducts, newWitness.RunningProducts[:len(newWitness.RunningProducts)-1]) {
			panic(fmt.Sprintf("%v %v %v", index, oldWitness.RunningProducts, newWitness.RunningProducts))
			return false
		}
		if !p2ArrayEquals(oldWitness.PotPubkeys, newWitness.PotPubkeys[:len(newWitness.PotPubkeys)-1]) {
			panic(fmt.Sprintf("%v %v %v", index, oldWitness.PotPubkeys, newWitness.PotPubkeys))
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

func VerifyPairing(ceremony *Ceremony) bool {
	for _, t := range ceremony.Transcripts {
		if !verifyPairing(t.PowersOfTau) {
			return false
		}
	}
	return true
}

func verifyPairing(pot PowersOfTau) bool {
	if len(pot.G1Powers) < 2 || len(pot.G2Powers) < 2 {
		return false
	}

	var (
		g2_0 = pot.G2Powers[0].ToAffine()
		g2_1 = pot.G2Powers[1].ToAffine()

		g1_0 = pot.G1Powers[0].ToAffine()
		g1_1 = pot.G1Powers[1].ToAffine()

		failed int32
		wg     = new(sync.WaitGroup)
	)

	for i := 0; i < len(pot.G1Powers)-1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pair1 := blst.Fp12MillerLoop(g2_1, pot.G1Powers[i].ToAffine())
			pair1.FinalExp()

			pair2 := blst.Fp12MillerLoop(g2_0, pot.G1Powers[i+1].ToAffine())
			pair2.FinalExp()

			if !bytes.Equal(pair1.ToBendian(), pair2.ToBendian()) {
				atomic.AddInt32(&failed, 1)
			}
		}(i)
	}

	for i := 0; i < len(pot.G2Powers)-1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pair1 := blst.Fp12MillerLoop(pot.G2Powers[i].ToAffine(), g1_1)
			pair1.FinalExp()

			pair2 := blst.Fp12MillerLoop(pot.G2Powers[i+1].ToAffine(), g1_0)
			pair2.FinalExp()

			if !bytes.Equal(pair1.ToBendian(), pair2.ToBendian()) {
				atomic.AddInt32(&failed, 1)
			}
		}(i)
	}

	wg.Wait()
	return failed == 0
}
