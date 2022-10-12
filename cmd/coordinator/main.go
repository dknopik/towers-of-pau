package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dknopik/towersofpau"
	"github.com/gorilla/mux"
)

func main() {
	if len(os.Args) < 2 {
		panic("invalid amount of args, need 2")
	}
	path := os.Args[1]
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("unable to open")
	}
	fmt.Println("Reading initial ceremony")
	ceremony, err := towersofpau.Deserialize(file)
	if err != nil {
		log.Fatal("unable to decode", err.Error())
	}
	err = os.Mkdir("history", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal("unable to create history dir", err.Error())
	}
	fmt.Println("Starting coordinator")
	coordinator := NewCoordinator(ceremony)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/participation", coordinator.RegisterParticipant).
		Methods("POST")
	router.HandleFunc("/participation/{ticket}", coordinator.RetrieveParticipant).
		Methods("GET")
	router.HandleFunc("/participation/{ticket}", coordinator.SubmitCeremony).
		Methods("POST")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
