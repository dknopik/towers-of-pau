package towersofpau

import (
	bls12381 "github.com/kilic/bls12-381"
)

var G1 = bls12381.NewG1()
var G2 = bls12381.NewG2()

func UpdateContribution(contribution *Contribution, secret []byte) error {
	if err := UpdatePowersOfTauFast(contribution, secret); err != nil {
		return err
	}

	if err := UpdateWitness(contribution, secret); err != nil {
		return err
	}

	return nil
}

func UpdatePowersOfTauFast(contribution *Contribution, secret []byte) error {
	sec := bls12381.NewFr().FromBytes(secret)
	scalar := sec
	scalars := make([]*bls12381.Fr, contribution.NumG1Powers)

	for i := 0; i < contribution.NumG1Powers; i++ {
		scalars[i] = scalar
		scalar.Mul(scalar, sec)
	}

	for i := 0; i < contribution.NumG1Powers; i++ {
		G1.MulScalar(contribution.PowersOfTau.G1Powers[i], contribution.PowersOfTau.G1Powers[i], scalar)

		if i < contribution.NumG2Powers {
			G2.MulScalar(contribution.PowersOfTau.G2Powers[i], contribution.PowersOfTau.G2Powers[i], scalar)
		}
	}
	return nil
}

func NewPotPubkey(secret []byte) *bls12381.PointG2 {
	sec := bls12381.NewFr().FromBytes(secret)
	potPubkey := G2.New()
	G2.MulScalar(potPubkey, &bls12381.G2One, sec)

	return potPubkey
}

func UpdateWitness(contribution *Contribution, secret []byte) error {
	sec := bls12381.NewFr().FromBytes(secret)
	G1.MulScalar(contribution.BLSSignature, contribution.BLSSignature, sec)

	contribution.PotPubkey = NewPotPubkey(secret)

	return nil
}
