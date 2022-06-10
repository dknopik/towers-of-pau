package towersofpau

import (
	"bytes"
	"encoding/hex"

	blst "github.com/supranational/blst/bindings/go"
)

type PowersOfTau struct {
	G1Powers []*blst.P1
	G2Powers []*blst.P2
}

type Witness struct {
	RunningProducts []*blst.P1
	PotPubkeys      blst.P2Affines
}

type Transcript struct {
	NumG1Powers int
	NumG2Powers int
	PowersOfTau PowersOfTau
	Witness     Witness
}

type Ceremony struct {
	Transcripts []*Transcript
}

type P1Affine blst.P1Affine

func (a *P1Affine) UnmarshalJSON(json []byte) error {
	json = bytes.Trim(json, "\"")
	hex.Decode()
	scalar := new(blst.Scalar).FromBEndian()
	return nil
}
