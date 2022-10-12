package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dknopik/towersofpau"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

const (
	participantTime     = 20
	coordinatorTime     = 60
	immediateStartDelay = 5
	pushbackDelay       = 10
	rounds              = 10
)

func NewCoordinator(initialCeremony *towersofpau.Ceremony) *Coordinator {
	return &Coordinator{
		slotByTicket: make(map[string]*slot),
		slots:        make([]*slot, 0),
		ceremony:     initialCeremony,
		maxRounds:    rounds,
	}
}

type Coordinator struct {
	mutex         sync.Mutex
	slotByTicket  map[string]*slot
	slots         []*slot
	currentSlot   int
	ceremony      *towersofpau.Ceremony
	ceremonyMutex sync.Mutex
	maxRounds     int
}

func (c *Coordinator) RegisterParticipant(rw http.ResponseWriter, req *http.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	setupResponse(&rw, req)

	slot := new(slot)
	slot.index = len(c.slots)
	// check if current slot has expired or is in processing
	for c.currentSlot < len(c.slots) {
		if c.slots[c.currentSlot].submitted || c.slots[c.currentSlot].deadline >= time.Now().Unix() {
			break
		} else {
			c.currentSlot++
		}
	}
	if c.currentSlot == slot.index {
		slot.start = time.Now().Unix() + immediateStartDelay
	} else {
		slot.start = c.slots[len(c.slots)-1].deadline + coordinatorTime
		if time.Now().Unix()+immediateStartDelay > slot.start {
			slot.start = time.Now().Unix() + immediateStartDelay
		}
	}
	slot.deadline = slot.start + participantTime
	slot.participantTicket = getTicket()
	c.slots = append(c.slots, slot)
	c.slotByTicket[slot.participantTicket] = slot
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
	fmt.Printf("Registered participant no. %v for %v\n", slot.index, time.Unix(slot.start, 0))
}

func getTicket() string {
	b := make([]byte, 32)
	l, err := rand.Read(b)
	if err != nil || l != 32 {
		panic("invalid randomness")
	}
	return common.Bytes2Hex(b)
}

func (c *Coordinator) RetrieveParticipant(rw http.ResponseWriter, req *http.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	setupResponse(&rw, req)
	ticket := mux.Vars(req)["ticket"]
	slot := c.slotByTicket[ticket]
	if slot == nil || slot.index < c.currentSlot {
		rw.WriteHeader(403)
		return
	}

	response := towersofpau.FetchResponse{
		Start:    slot.start,
		Deadline: slot.deadline,
		Ceremony: nil,
	}

	// check if current slot has expired or is in processing
	for c.currentSlot < slot.index {
		if c.slots[c.currentSlot].submitted || c.slots[c.currentSlot].deadline >= time.Now().Unix() {
			break
		} else {
			c.currentSlot++
		}
	}

	if c.currentSlot == slot.index {
		if slot.deadline < time.Now().Unix() {
			c.currentSlot++
			rw.WriteHeader(403)
			return
		} else {
			jsonceremony, err := towersofpau.SerializeJSONCeremony(c.ceremony)
			if err != nil {
				rw.WriteHeader(500)
				return
			}
			response.Ceremony = &jsonceremony
		}
	} else if c.currentSlot < slot.index && slot.start <= time.Now().Unix()+1 {
		for _, slot := range c.slots[c.currentSlot+1:] {
			slot.start += pushbackDelay
			slot.deadline += pushbackDelay
		}
		response.Start = slot.start
		response.Deadline = slot.deadline
	}

	resp, err := json.Marshal(response)
	if err != nil {
		rw.WriteHeader(500)
		return
	}
	rw.Write(resp)
	fmt.Printf("Participant no. %v retrieved ceremony\n", slot.index)
}

func (c *Coordinator) SubmitCeremony(rw http.ResponseWriter, req *http.Request) {
	c.mutex.Lock()
	setupResponse(&rw, req)
	ticket := mux.Vars(req)["ticket"]
	slot := c.slotByTicket[ticket]
	if slot == nil || slot.index != c.currentSlot {
		c.mutex.Unlock()
		rw.WriteHeader(403)
		return
	}
	fmt.Printf("Received submission from %v\n", slot.index)
	slot.submitted = true
	c.mutex.Unlock()

	newCeremony, err := towersofpau.Deserialize(req.Body)
	if err != nil {
		c.currentSlot++
		rw.WriteHeader(400)
		return
	}

	c.ceremonyMutex.Lock()
	defer c.ceremonyMutex.Unlock()
	oldCeremony := c.ceremony
	fmt.Printf("Verifying submission from %v\n", slot.index)
	start := time.Now()
	if err := towersofpau.VerifySubmission(oldCeremony, newCeremony); err != nil {
		c.currentSlot++
		fmt.Printf("Submission verification from %v failed: %v\n", slot.index, err)
		rw.WriteHeader(400)
		return
	}
	fmt.Printf("Submission verified successfully in %v\n", time.Since(start))
	// Ceremony was valid, store it
	c.ceremony = newCeremony
	rw.WriteHeader(200)

	c.currentSlot++

	file, err := os.Create(fmt.Sprintf("history/%d.json", slot.index))
	if err != nil {
		return
	}
	towersofpau.Serialize(file, newCeremony)
	return
}

type slot struct {
	index             int
	start             int64
	deadline          int64
	participantTicket string
	submitted         bool
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
