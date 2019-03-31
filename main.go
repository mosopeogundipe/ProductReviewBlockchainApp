package main

import (
	"cs686-blockchain-p3-mosopeogundipe/p3"
	"log"
	"net/http"
	"os"
)

func main() {
	//data.TestPeerListRebalance()
	router := p3.NewRouter()
	if len(os.Args) >= 1 {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	}
}
