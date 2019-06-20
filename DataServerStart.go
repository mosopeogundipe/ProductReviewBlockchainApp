package main

import (
	"cs686-blockchain-p3-mosopeogundipe/p5"
	"log"
	"net/http"
)

func main() {
	router := p5.NewRouter()
	log.Fatal(http.ListenAndServe(":"+"7700", router)) //use static port 7700 for this server
}
