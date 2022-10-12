package towersofpau

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	bls12381 "github.com/kilic/bls12-381"
)

var G1 = bls12381.NewG1()
var G2 = bls12381.NewG2()

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
	sec := bls12381.NewFr().FromBytes(secret)
	scalar := &(*sec)
	for i := 0; i < transcript.NumG1Powers; i++ {
		G1.MulScalar(transcript.PowersOfTau.G1Powers[i], transcript.PowersOfTau.G1Powers[i], scalar)
		if i < transcript.NumG2Powers {
			G2.MulScalar(transcript.PowersOfTau.G2Powers[i], transcript.PowersOfTau.G2Powers[i], scalar)
		}
		scalar.Mul(scalar, sec)
	}
	return nil
}

func UpdatePowersOfTauFast(transcript *Transcript, secret []byte) error {
	sec := bls12381.NewFr().FromBytes(secret)
	scalar := &(*sec)
	scalars := make([]*bls12381.Fr, transcript.NumG1Powers)
	for i := 0; i < transcript.NumG1Powers; i++ {
		scalars[i] = &(*scalar)
		scalar.Mul(scalar, sec)
	}

	wg := new(sync.WaitGroup)
	wg.Add(transcript.NumG1Powers)
	for i := 0; i < transcript.NumG1Powers; i++ {
		go func(i int) {
			defer wg.Done()
			G1.MulScalar(transcript.PowersOfTau.G1Powers[i], transcript.PowersOfTau.G1Powers[i], scalar)
			if i < transcript.NumG2Powers {
				G2.MulScalar(transcript.PowersOfTau.G2Powers[i], transcript.PowersOfTau.G2Powers[i], scalar)
			}
		}(i)
	}
	wg.Wait()
	return nil
}

// UpdateWitness updates the witness with our secret.
func UpdateWitness(transcript *Transcript, secret []byte) error {
	newProduct := &(*transcript.Witness.RunningProducts[len(transcript.Witness.RunningProducts)-1])
	sec := bls12381.NewFr().FromBytes(secret)
	G1.MulScalar(newProduct, newProduct, sec)
	transcript.Witness.RunningProducts = append(transcript.Witness.RunningProducts, newProduct)
	newPk := G2.New()
	G2.MulScalar(newPk, &bls12381.G2One, sec)
	if newPk == nil {
		return errors.New("invalid pk")
	}
	transcript.Witness.PotPubkeys = append(transcript.Witness.PotPubkeys, newPk)
	return nil
}

func createRandom() *big.Int {
	for i := 0; i < 1000000; i++ {
		b := make([]byte, 32)
		n, err := rand.Read(b)
		if err != nil || n != 32 {
			panic("could not get good randomness")
		}

		return new(big.Int).SetBytes(b)

	}
	panic("could not find secret in 1 million tries")
}

// SubgroupChecks verifies that a ceremony looks correctly
func SubgroupChecksParticipant(ceremony *Ceremony) bool {
	for _, transcript := range ceremony.Transcripts {
		for _, p := range transcript.PowersOfTau.G1Powers {
			if !G1.IsOnCurve(p) {
				return false
			}
		}
		for _, p := range transcript.PowersOfTau.G2Powers {
			if !G2.IsOnCurve(p) {
				return false
			}
		}
		for _, p := range transcript.Witness.RunningProducts {
			if !G1.IsOnCurve(p) {
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
			if !G1.IsOnCurve(p) {
				return false
			}
		}
		for _, p := range transcript.PowersOfTau.G2Powers {
			if !G2.IsOnCurve(p) {
				return false
			}
		}
		for _, p := range transcript.Witness.RunningProducts {
			if !G1.IsOnCurve(p) {
				return false
			}
		}
		for _, p := range transcript.Witness.PotPubkeys {
			if !G2.IsOnCurve(p) {
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
			// TODO reenable this check1
			// if !p.IsInfinite() { return false}
		}
	}
	return true
}

func PubkeyUniquenessCheck(ceremony *Ceremony) bool {
	keys := make(map[*bls12381.PointG2]struct{}, 0)
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

func p1ArrayEquals(p1, p2 []*bls12381.PointG1) bool {
	if len(p1) != len(p2) {
		return false
	}
	for idx := range p1 {
		if !G1.Equal(p1[idx], p2[idx]) {
			return false
		}
	}
	return true
}

func p2ArrayEquals(p1, p2 []*bls12381.PointG2) bool {
	if len(p1) != len(p2) {
		return false
	}
	for idx := range p1 {
		if !G2.Equal(p1[idx], p2[idx]) {
			return false
		}
	}
	return true
}

func VerifyPairing(ceremony *Ceremony) bool {
	for _, t := range ceremony.Transcripts {
		if !verifyPairing(t) {
			return false
		}
	}
	return true
}

func verifyPairing(t *Transcript) bool {
	if len(t.Witness.PotPubkeys) != len(t.Witness.RunningProducts) ||
		len(t.PowersOfTau.G1Powers) < 2 || len(t.PowersOfTau.G2Powers) < 2 {
		return false
	}

	var (
		g2_0 = t.PowersOfTau.G2Powers[0]
		g2_1 = t.PowersOfTau.G2Powers[1]

		g1_0 = t.PowersOfTau.G1Powers[0]
		g1_1 = t.PowersOfTau.G1Powers[1]

		failed int32
		wg     = new(sync.WaitGroup)
	)

	wg.Add(len(t.PowersOfTau.G1Powers) - 1)
	for i := 0; i < len(t.PowersOfTau.G1Powers)-1; i++ {
		go func(i int) {
			defer wg.Done()
			engine := bls12381.NewEngine()
			engine.AddPair(t.PowersOfTau.G1Powers[i], g2_1)
			engine.AddPair(t.PowersOfTau.G1Powers[i+1], g2_0)
			if !engine.Check() {
				atomic.AddInt32(&failed, 1)
			}
		}(i)
	}
	wg.Wait()

	wg.Add(len(t.PowersOfTau.G2Powers) - 1)
	for i := 0; i < len(t.PowersOfTau.G2Powers)-1; i++ {
		go func(i int) {
			defer wg.Done()
			engine := bls12381.NewEngine()
			engine.AddPair(g1_1, t.PowersOfTau.G2Powers[i])
			engine.AddPair(g1_0, t.PowersOfTau.G2Powers[i+1])
			if !engine.Check() {
				atomic.AddInt32(&failed, 1)
			}
		}(i)
	}
	wg.Wait()

	p2_g := G2.One()
	wg.Add(len(t.Witness.RunningProducts) - 1)
	for i := 0; i < len(t.Witness.RunningProducts)-1; i++ {
		go func(i int) {
			defer wg.Done()
			engine := bls12381.NewEngine()
			engine.AddPair(t.Witness.RunningProducts[i], t.Witness.PotPubkeys[i+1])
			engine.AddPair(t.Witness.RunningProducts[i+1], p2_g)
			if !engine.Check() {
				atomic.AddInt32(&failed, 1)
			}
		}(i)
	}
	wg.Wait()
	return failed == 0
}
