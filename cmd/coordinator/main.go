package coordinator

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var slotByTicket = make(map[string]*slot)
var slots = make([]*slot, 0)
var currentSlot = 0

const (
	participantTime     = 20
	coordinatorTime     = 120
	immediateStartDelay = 5
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/participation", registerParticipant).
		Methods("POST")
	router.HandleFunc("/participation/{ticket}", retrieveParticipant).
		Methods("GET")
	router.HandleFunc("/participation/{ticket}", submitCeremony).
		Methods("POST")
	err := http.ListenAndServe(":2016", router)
	if err != nil {
		log.Fatal(err)
	}
}

func registerParticipant(rw http.ResponseWriter, req *http.Request) {
	slot := new(slot)
	slot.index = len(slots)
	if currentSlot == slot.index {
	}
}

func retrieveParticipant(http.ResponseWriter, *http.Request) {

}

func submitCeremony(http.ResponseWriter, *http.Request) {

}

type slot struct {
	index             int
	start             int
	deadline          int
	participantTicket string
}
