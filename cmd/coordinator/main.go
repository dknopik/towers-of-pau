package main

import (
	"encoding/json"
	"github.com/dknopik/towersofpau"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var mutex sync.Mutex
var slotByTicket = make(map[string]*slot)
var slots = make([]*slot, 0)
var currentSlot = 0
var ceremony *towersofpau.Ceremony = nil

const (
	participantTime     = 20
	coordinatorTime     = 120
	immediateStartDelay = 5
	pushbackDelay       = 10
)

func main() {
	rand.Seed(time.Now().UnixNano())
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		log.Fatal("unable to open")
	}
	ceremony, err = towersofpau.Deserialize(file)
	if err != nil {
		log.Fatal("unable to decode", err.Error())
	}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/participation", registerParticipant).
		Methods("POST")
	router.HandleFunc("/participation/{ticket}", retrieveParticipant).
		Methods("GET")
	router.HandleFunc("/participation/{ticket}", submitCeremony).
		Methods("POST")
	err = http.ListenAndServe(":2016", router)
	if err != nil {
		log.Fatal(err)
	}
}

func registerParticipant(rw http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	slot := new(slot)
	slot.index = len(slots)
	if currentSlot == slot.index {
		slot.start = time.Now().Unix() + immediateStartDelay
	} else {
		slot.start = slots[len(slots)-1].deadline + coordinatorTime
	}
	slot.deadline = slot.start + participantTime
	slot.participantTicket = getTicket()
	slotByTicket[slot.participantTicket] = slot
	resp, err := json.Marshal(towersofpau.RegistrationResponse{
		Start:    slot.start,
		Deadline: slot.deadline,
		Ticket:   slot.participantTicket,
	})
	if err != nil {
		rw.WriteHeader(500)
		return
	}
	rw.Write(resp)
}

const ticketBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func getTicket() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = ticketBytes[rand.Intn(len(ticketBytes))]
	}
	return string(b)
}

func retrieveParticipant(rw http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	ticket := mux.Vars(req)["ticket"]
	slot := slotByTicket[ticket]
	if slot == nil || slot.index < currentSlot {
		rw.WriteHeader(403)
		return
	}

	response := towersofpau.FetchResponse{
		Start:    slot.start,
		Deadline: slot.deadline,
		Ceremony: nil,
	}

	// check if current slot has expired or is in processing
	if currentSlot < slot.index {

	}

	if currentSlot == slot.index {
		jsonceremony, err := towersofpau.SerializeJSONCeremony(ceremony)
		if err != nil {
			rw.WriteHeader(500)
			return
		}
		response.Ceremony = &jsonceremony
	}

	if currentSlot < slot.index && slot.start < time.Now().Unix() {
		for _, slot := range slots[currentSlot+1:] {
			slot.start += pushbackDelay
			slot.deadline += pushbackDelay
		}
	}

	resp, err := json.Marshal(response)
	if err != nil {
		rw.WriteHeader(500)
		return
	}
	rw.Write(resp)
}

func submitCeremony(rw http.ResponseWriter, req *http.Request) {

}

type slot struct {
	index             int
	start             int64
	deadline          int64
	participantTicket string
	submitted         bool
}
