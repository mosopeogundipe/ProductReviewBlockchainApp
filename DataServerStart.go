package main

import (
	"cs686-blockchain-p3-mosopeogundipe/p5"
	"log"
	"net/http"
)

func main() {
	//data.TestPeerListRebalance()
	router := p5.NewRouter()
	log.Fatal(http.ListenAndServe(":"+"7700", router)) //use static port 7700 for this server
	//if len(os.Args) >= 1 {
	//	log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	//} else {
	//	log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	//}
}
