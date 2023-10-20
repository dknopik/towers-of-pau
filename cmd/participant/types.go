package main

import "fmt"

type CeremonyStatus struct {
	LobbySize        int32  `json:"lobby_size"`
	NumContributions int32  `json:"num_contributions"`
	SequencerAddress string `json:"sequencer_address"`
}

func (c *CeremonyStatus) String() string {
	return fmt.Sprintf("LobbySize: %v, NumContributions: %v, Sequencer: %v", c.LobbySize, c.NumContributions, c.SequencerAddress)
}

type Participant struct {
	IdType string
	Id     string
}

type AuthResponse struct {
	Participant Participant
	Token       string
}

type ErrorStruct struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}
