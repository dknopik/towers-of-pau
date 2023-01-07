package towersofpau

import (
	"errors"
	"sync"

	bls12381 "github.com/kilic/bls12-381"
)

var G1 = bls12381.NewG1()
var G2 = bls12381.NewG2()

func UpdateTranscript(transcript *Transcript, secret []byte) error {
	if err := UpdatePowersOfTauFast(transcript, secret); err != nil {
		return err
	}

	if err := UpdateWitness(transcript, secret); err != nil {
		return err
	}

	return nil
}

func UpdatePowersOfTauFast(transcript *Transcript, secret []byte) error {
	sec := bls12381.NewFr().FromBytes(secret)
	scalar := sec
	scalars := make([]*bls12381.Fr, transcript.NumG1Powers)

	for i := 0; i < transcript.NumG1Powers; i++ {
		scalars[i] = scalar
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

func UpdateWitness(transcript *Transcript, secret []byte) error {
	newProduct := transcript.Witness.RunningProducts[len(transcript.Witness.RunningProducts)-1]
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
