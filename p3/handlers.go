package p3

import (
	"bytes"
	"cs686-blockchain-p3-mosopeogundipe/p2"
	"cs686-blockchain-p3-mosopeogundipe/p3/data"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:6688"
var HEART_BEAT_API_SUFFIX = "/heartbeat/receive"
var UPLOAD_BLOCK_SUFFIX = "/block"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool //using this to indicate if it's started sending heartbeat
var hostname string
var port_ string

func init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	ifStarted = false
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	//Register()
	SELF_ADDR = r.Host
	_, port, _ := net.SplitHostPort(r.Host)
	if !ifStarted {
		hostname = r.Host
		port_ = port
		portInt, _ := strconv.Atoi(port)
		Peers = data.NewPeerList(int32(portInt), 32) //--> Uncomment this line! Get ID from command line here
		fmt.Println("Port is: ", port)
		fmt.Println("Port: ", r.URL.Port())
		//fmt.Println("Url: ", host)
		if port == "6688" { //if it's primary node
			fmt.Println("Is Primary Node")
			go StartHeartBeat()
		} else {
			Download()
			go StartHeartBeat()
		}
		ifStarted = true
	}

}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server, get an ID
func Register() {
	resp, err := http.Get(REGISTER_SERVER)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var strId = string(body)
	id, err := strconv.Atoi(strId)
	if err != nil {
		fmt.Println(err)
	}
	Peers = data.NewPeerList(int32(id), 32)
	//return id
}

// Download blockchain from TA server
func Download() {
	if !ifStarted {
		resp, err := http.Get(BC_DOWNLOAD_SERVER + "?host=" + hostname + "&id=" + port_)
		if err != nil {
			log.Fatal("Error in Download: ", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			blockChainJson, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
				return
			}
			SBC.UpdateEntireBlockChain(string(blockChainJson))
		} else {
			log.Println("No blockchain returned from primary node!")
		}
	}
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	log.Println("Entered Upload")
	q := r.URL.Query()
	host := q["host"][0]
	portStr := q["id"][0]
	log.Println("In Upload host: ", host, "Port: ", portStr)
	//_, portStr, _ := net.SplitHostPort(r.Host)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(err, "Upload")
	}

	if host != SELF_ADDR { //do not add node's self address to peerMap
		Peers.Add("http://"+host, int32(port))
	}
	//Peers.Add("http://" + host, int32(port))
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		log.Fatal(err, " Wowowow")
		w.WriteHeader(404)
		fmt.Fprint(w, "")
		log.Println("Block Not Found ! Leaving Upload")
	} else {
		//Peers.Add("http://" + host, int32(port))
		w.WriteHeader(200)
		fmt.Fprint(w, blockChainJson)
	}
	log.Println("Leaving Upload")
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	log.Println("Entered UploadBlock")
	splitPath := strings.Split(r.URL.Path, "/") //First element of array is an empty string
	heightStr := splitPath[1]
	hash := splitPath[2]
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		log.Fatal("Error in UploadBlock: strconv", err)
	}
	block, boolean := SBC.GetBlock(int32(height), hash)
	if !boolean {
		log.Fatal("In UploadBlock: block not found!")
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	fmt.Fprint(w, block.EncodeToJSON())
	log.Println("Leaving UploadBlock")
}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	log.Println("In HeartBeatReceive")
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err, "Error in reading post request: HeartBeatReceive ")
	}
	var heartBeatData data.HeartBeatData
	json.Unmarshal([]byte(request), &heartBeatData)
	log.Println("HeartBeatData: ", heartBeatData)
	addr := heartBeatData.Addr
	if addr != SELF_ADDR { //do not add node's self address to peerMap
		Peers.Add("http://"+addr, heartBeatData.Id)
	}
	//Peers.Add("http://" + addr, heartBeatData.Id)
	Peers.InjectPeerMapJson(heartBeatData.PeerMapJson, heartBeatData.Addr)
	if heartBeatData.IfNewBlock {
		var block p2.Block
		block = block.DecodeFromJson(heartBeatData.BlockJson)
		//json.Unmarshal([]byte(heartBeatData.BlockJson), block)
		log.Println("In HeartBeatReceive. BlockJson: ", heartBeatData.BlockJson)
		if !SBC.CheckParentHash(block) {
			//Parent Block doesn't exist, so fetch it
			parentHeight := block.Header.Height - 1
			AskForBlock(parentHeight, block.Header.ParentHash)
		}
		SBC.Insert(block)
		heartBeatData.Hops = heartBeatData.Hops - 1
		if heartBeatData.Hops >= 0 {
			ForwardHeartBeat(heartBeatData)
		}
	} else {
		//Do nothing!
		log.Println("HeartBeatReceive: It's not a new block, so nothing is done")
	}
	w.WriteHeader(200)
	fmt.Println("Leaving HeartBeatReceive")
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	log.Println("Entered AskForBlock. Hash: ", hash)
	isBlockFound := false
	var block p2.Block
	for !isBlockFound {
		for k, _ := range Peers.Copy() {
			//call send heart beat here
			peerUrl := k + UPLOAD_BLOCK_SUFFIX + "/" + string(height) + "/" + hash
			resp, err := http.Get(peerUrl)
			if err != nil {
				log.Fatal("Error in AskForBlock: UploadBlock", err)
				return
			}
			defer resp.Body.Close()
			blockJson, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Block Not Found! Leaving AskForBlock")
				log.Fatalln(err)
				return
			}
			if resp.StatusCode == 200 {
				isBlockFound = true
				//json.Unmarshal([]byte(blockJson), block)
				block = block.DecodeFromJson(string(blockJson))
				SBC.Insert(block)
				break
			}
		}
		//log.Println("Leaving AskForBlock")
	}
}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	if heartBeatData.Hops >= 0 {
		fmt.Println("Peers", Peers)
		Peers.Rebalance() //Rebalance peerList before sending
		heartBeatJson, _ := json.Marshal(heartBeatData)
		for k, v := range Peers.Copy() {
			fmt.Println("Key: ", k, "Value: ", v)
			//call send heart beat here
			peerUrl := k + HEART_BEAT_API_SUFFIX
			httpPost(peerUrl, string(heartBeatJson))
		}
	}
}

func httpPost(url string, jsonBody string) {
	log.Println("Entered POST")
	var jsonStr = []byte(jsonBody)
	//client := http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("charset", "UTF-8")
	//resp, _ := client.Do(req)
	//defer resp.Body.Close()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("Error in POST. Status is:", resp.Status)
	}
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
	log.Println("Finished POST")
}

func StartHeartBeat() {
	for { //infinite for loo
		fmt.Println("In StartHeartBeat")
		time.Sleep(5 * time.Second)
		peerMapJson, _ := Peers.PeerMapToJson()
		heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR)
		//heartBeatData.Hops -= 1
		ForwardHeartBeat(heartBeatData)
	}
}
