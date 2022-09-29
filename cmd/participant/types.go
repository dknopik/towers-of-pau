package main

type QueueStatus string

const (
	notQueued         QueueStatus = "notQueued"
	queued            QueueStatus = "queued"
	readyToContribute QueueStatus = "readyToContribute"
	contributing      QueueStatus = "contributing"
	expired           QueueStatus = "expired"
)

type CeremonyStatus struct {
}

type ParticipantQueueStatus struct {
	QueuePosition   int32
	NextCheckinTime int32
	QueueStatus     QueueStatus
	Id              string
}

type Participant struct {
	IdType string
	Id     string
}

type AuthResponse struct {
	Participant Participant
	Token       string
}
