package main

import (
	"github.com/dknopik/towersofpau"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	file, err := os.Open("initialCeremony.json")
	if err != nil {
		log.Fatal("unable to open")
	}
	ceremony, err := towersofpau.Deserialize(file)
	if err != nil {
		log.Fatal("unable to decode", err.Error())
	}
	coordinator := NewCoordinator(ceremony)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/participation", coordinator.RegisterParticipant).
		Methods("POST")
	router.HandleFunc("/participation/{ticket}", coordinator.RetrieveParticipant).
		Methods("GET")
	router.HandleFunc("/participation/{ticket}", coordinator.SubmitCeremony).
		Methods("POST")
	err = http.ListenAndServe(":2016", router)
	if err != nil {
		log.Fatal(err)
	}
}
