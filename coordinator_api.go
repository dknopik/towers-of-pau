package towersofpau

type RegistrationResponse struct {
	Start    int64
	Deadline int64
	Ticket   string
}

type FetchResponse struct {
	Start    int64
	Deadline int64
	Ceremony *JSONCeremony
}
