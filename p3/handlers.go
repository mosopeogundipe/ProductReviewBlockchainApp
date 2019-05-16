package p3

import (
	"bytes"
	"cs686-blockchain-p3-mosopeogundipe/p1"
	"cs686-blockchain-p3-mosopeogundipe/p2"
	"cs686-blockchain-p3-mosopeogundipe/p3/data"
	data2 "cs686-blockchain-p3-mosopeogundipe/p5/data"
	"cs686-blockchain-p3-mosopeogundipe/p5/logic"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var MIDDLE_LAYER_SERVER = "http://localhost:7700" //IP and port of middle layer server
var SELF_ADDR = "http://localhost:6688"
var HEART_BEAT_API_SUFFIX = "/heartbeat/receive"
var UPLOAD_BLOCK_SUFFIX = "/block"
var INTERVAL = 5
var MUTEX = &sync.Mutex{}

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool //using this to indicate if it's started sending heartbeat
var hostname string
var port_ string
var proposedParentHash string
var transactionQueue []data2.Transaction

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
		Peers = data.NewPeerList(int32(portInt), 32)
		fmt.Println("Port is: ", port)
		fmt.Println("Port: ", r.URL.Port())
		RegisterInMiddleLayer()
		if port == "6688" { //if it's primary node
			fmt.Println("Is Primary Node")
			go StartHeartBeat()
		} else {
			Download()
			go StartHeartBeat()
		}
		go StartTryingNonces()
		go SBC.FinishInsert() //in this thread, insert blocks from queue into blockchain
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
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
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
			log.Println("Error in Download: ", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			blockChainJson, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return
			}
			SBC.UpdateEntireBlockChain(string(blockChainJson))
		} else {
			log.Println("No blockchain returned from primary node!")
		}
	}
}

func RegisterInMiddleLayer() {
	resp, err := http.Get(MIDDLE_LAYER_SERVER + "/miner/register" + "?host=" + hostname + "&id=" + port_)
	if err != nil {
		log.Println("Error in RegisterMiner: ", err)
		//return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println(hostname + " Registered Successfully")
	} else {
		for resp.StatusCode != 200 { //keep retrying until miner succesfully registers
			log.Println("Registration Unsuccessful! Retrying...: ", err)
			resp, err = http.Get(MIDDLE_LAYER_SERVER + "/registerminer" + "?host=" + hostname + "&id=" + port_)
			if err != nil {
				log.Println("Error in RegisterMiner: ", err)
				//return
			}
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
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err, "Upload")
	}

	if host != SELF_ADDR { //do not add node's self address to peerMap
		Peers.Add("http://"+host, int32(port))
	}
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		log.Println(err, " Wowowow")
		w.WriteHeader(404)
		fmt.Fprint(w, "")
		log.Println("Block Not Found ! Leaving Upload")
	} else {
		w.WriteHeader(200)
		fmt.Fprint(w, blockChainJson)
	}
	log.Println("Leaving Upload")
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	log.Println("Entered UploadBlock")
	log.Println("In UploadBlock. Url Path: ", r.URL.Path)
	splitPath := strings.Split(r.URL.Path, "/") //First element of array is an empty string
	heightStr := splitPath[2]
	hash := splitPath[3]
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		log.Println("Error in UploadBlock: strconv", err)
	}
	block, boolean := SBC.GetBlock(int32(height), hash)
	if !boolean {
		log.Println("In UploadBlock: block not found!")
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
		log.Println(err, "Error in reading post request: HeartBeatReceive ")
	}
	var heartBeatData data.HeartBeatData
	json.Unmarshal([]byte(request), &heartBeatData)
	log.Println("HeartBeatData: ", heartBeatData)
	addr := heartBeatData.Addr
	if addr != SELF_ADDR { //do not add node's self address to peerMap
		Peers.Add("http://"+addr, heartBeatData.Id)
	}
	Peers.InjectPeerMapJson(heartBeatData.PeerMapJson, heartBeatData.Addr)
	if heartBeatData.IfNewBlock {
		var block p2.Block
		block = block.DecodeFromJson(heartBeatData.BlockJson)
		value := sha3.Sum256([]byte(block.Header.ParentHash + block.Header.Nonce + block.Value.GetRoot()))
		valueStr := hex.EncodeToString(value[:])            // Do we need to convert to string to verify 0s OR should we just check 'value' byte array for successive 0s?
		if strings.HasPrefix(valueStr, data.NONCE_PREFIX) { //only accept block that passes verification
			log.Println("accepting block with hash: ", block.Header.Hash)
			if block.Header.ParentHash == proposedParentHash {
				data.IS_PARENT_USED_ALREADY = true
			}
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
			log.Println("Block doesn't pass verification. Skipping insert. Hash: ", block.Header.Hash)
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
	if height == 0 { //stop looking for parent height once you reach genesis block (block at height 0)
		log.Println("reached end of recursion in AskForBlock")
		return
	}
	log.Println("Entered AskForBlock. Height:", height, " Hash: ", hash)
	//isBlockFound := false
	var block p2.Block
	//for !isBlockFound {
	for k, _ := range Peers.Copy() {
		//call send heart beat here
		heightStr := strconv.Itoa(int(height))
		peerUrl := k + UPLOAD_BLOCK_SUFFIX + "/" + heightStr + "/" + hash
		resp, err := http.Get(peerUrl)
		if err != nil {
			log.Println("Error in AskForBlock: UploadBlock", err)
			//return
		}
		defer resp.Body.Close()
		blockJson, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Block Not Found! Leaving AskForBlock")
			log.Println(err)
			//return
		}
		if resp.StatusCode == 200 {
			block = block.DecodeFromJson(string(blockJson))
			if block.Header.ParentHash == proposedParentHash {
				data.IS_PARENT_USED_ALREADY = true
			}
			if !SBC.CheckParentHash(block) {
				//Parent Block doesn't exist, so fetch it
				parentHeight := block.Header.Height - 1
				AskForBlock(parentHeight, block.Header.ParentHash)
			}
			SBC.Insert(block)
			break
		}
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
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("Error in POST. Status is:", resp.Status)
	}
	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	log.Println("Finished POST")
}

func StartHeartBeat() {
	for { //infinite loop - stops when program is closed
		fmt.Println("In StartHeartBeat")
		interval := rand.Intn(5) + 5
		intervalSecs, _ := time.ParseDuration(strconv.Itoa(interval) + "s")
		time.Sleep(intervalSecs)
		peerMapJson, _ := Peers.PeerMapToJson()
		heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, p2.Block{}, false)
		ForwardHeartBeat(heartBeatData)
	}
}

func StartTryingNonces() {
	timeStamp := 0
	for {
		_, isGenesisBlockCreated := SBC.Get(0)
		var block p2.Block
		if !isGenesisBlockCreated {
			log.Println("Entered Create Genesis Block...")
			mpt := data.CreateRandomMpt() //only create rendom mpt for genesis block
			block = SBC.GenBlock(mpt)
			peerMapJson, _ := Peers.PeerMapToJson()
			heartbeat := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, block, true)
			ForwardHeartBeat(heartbeat)
			continue
		}
		//log.Println("timeStamp: ", timeStamp, "current time: ", time.Now().Second())
		if time.Now().Second()-timeStamp > INTERVAL { //will now generate new block every 5 seconds
			//for non-genesis blocks, make mpt from current transactions in pool
			log.Println("Entered Create Other Blocks. Transaction Count: ", len(transactionQueue))
			if len(transactionQueue) > 0 {
				log.Println("Transactions exist...")
				mpt := p1.MerklePatriciaTrie{}
				mpt.Initial()
				//MUTEX.Lock()
				for i := 0; i < len(transactionQueue); i++ {
					sum := sha3.Sum256([]byte(transactionQueue[0].PublicKey))
					hash := hex.EncodeToString(sum[:])
					transactionJson, _ := json.Marshal(transactionQueue[0])
					mpt.Insert(hash, string(transactionJson))
					// Dequeue current transaction from pool
					transactionQueue[0] = data2.Transaction{}
					transactionQueue = transactionQueue[1:]
				}
				//MUTEX.Unlock()
				parentBlock := SBC.GetLatestBlocks()[0]
				proposedParentHash = parentBlock.Header.Hash
				nonce := data.GetNonce(parentBlock, mpt.GetRoot())
				log.Println("got nonce!")
				if data.IS_PARENT_USED_ALREADY { //checking if proposed parent is already used as a parent, just before I insert...
					log.Println("Hash: ", parentBlock.Header.Hash, " is already used to create a block by another node")
					data.IS_PARENT_USED_ALREADY = false //reset checker value to default
					continue
				}
				block = p2.Initial(parentBlock.Header.Hash, parentBlock.Header.Height, mpt)
				block.Header.Nonce = nonce
				//data.SetValues(true, block)
				peerMapJson, _ := Peers.PeerMapToJson()
				heartbeat := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, block, true)
				ForwardHeartBeat(heartbeat)
			} else {
				log.Println("Could not create new block : No transactions in Queue!")
			}
			timeStamp = 0
		}
	}
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	var value string = ""
	for i := range SBC.GetLatestBlocks() {
		value += "Chain " + strconv.Itoa(i+1) + "\n"
		block := SBC.GetLatestBlocks()[i]
		value += PrintBlock(block)
		height := block.Header.Height
		log.Println("In Canonical ", "Chain "+string(i+1), " start height: ", height)
		for height > 1 {
			parent, exists := SBC.GetBlock(height-1, block.Header.ParentHash)
			if exists {
				value += PrintBlock(parent)
			} else {
				log.Println("Parent Hash: ", parent.Header.Hash, " doesn't exist in blockchain")
			}
			block = parent
			height--
		}
	}
	fmt.Fprintf(w, value)
}

func PrintBlock(block p2.Block) string {
	return "height = " + strconv.Itoa(int(block.Header.Height)) + ", " + "timestamp = " + strconv.Itoa(int(block.Header.Timestamp)) + ", " + "parentHash = " + block.Header.ParentHash + ", " + "size = " + strconv.Itoa(int(block.Header.Size)) + "\n"
}

// Receive transactions from middle layer, validate and store them in transaction pool if they are valid
func TransactionReceive(w http.ResponseWriter, r *http.Request) {
	var transaction data2.Transaction
	log.Println("In TransactionReceive")
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err, "Error in reading post request: TransactionReceive ")
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal([]byte(request), &transaction)
	if err != nil {
		log.Println(err, "Error in json unmarshal: TransactionReceive ")
		w.WriteHeader(500)
		return
	}
	q := r.URL.RawQuery
	log.Println("raw query: ", q)
	signatureStr := strings.TrimPrefix(q, "Signature=")
	originallySignedMsg := data2.Transaction{TransactionID: "", PublicKey: "", ReviewObj: transaction.ReviewObj} //create object with exact same contents as when message was signed
	originallySignedMsgJson, _ := json.Marshal(originallySignedMsg)
	//log.Println("Signature: ", signatureStr)
	log.Println("JSON message: ", string(originallySignedMsgJson))
	signature, err := base64.StdEncoding.DecodeString(signatureStr) //convert signature string back to byte array
	if err != nil {
		log.Println(err, "Error in decoding signature to byte array: TransactionReceive ")
		w.WriteHeader(500)
		return
	}
	isSignatureVerified := logic.VerifyPrivateKeySignature([]byte(string(originallySignedMsgJson)), signature, []byte(transaction.PublicKey))
	if isSignatureVerified { //add to pool if signature is verified
		log.Println("Signature is Verified!")
		log.Println("Transaction pool size: ", len(transactionQueue))
		if len(transactionQueue) == 0 {
			MUTEX.Lock()
			transactionQueue = append(transactionQueue, transaction)
			MUTEX.Unlock()
			w.WriteHeader(200)
		} else {
			for tx := range transactionQueue {
				if transactionQueue[tx] == transaction {
					w.WriteHeader(200)
					w.Write([]byte("Transaction is currently in miner's queue. Can't add same transaction multiple times"))
					return
				}
			}
			MUTEX.Lock()
			transactionQueue = append(transactionQueue, transaction)
			MUTEX.Unlock()
			w.WriteHeader(200)
		}
	} else {
		log.Println("Transaction Invalid. Public-Private key mismatch")
		w.WriteHeader(400)
		w.Write([]byte("Transaction Invalid. Public-Private key mismatch"))
	}
}

func FindReviewsByPublicKey(w http.ResponseWriter, r *http.Request) {
	log.Println("Entered FindReviewsByPublicKey")
	var userUpload data2.UserWebUpload
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err, "Error in reading post request: FindReviewsByPublicKey ")
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal([]byte(request), &userUpload)
	if err != nil {
		log.Println(err, "Error in json unmarshal: FindReviewsByPublicKey ")
		w.WriteHeader(500)
		return
	}
	if userUpload.PublicKey == "" {
		log.Println(err, "No public key given")
		w.WriteHeader(500)
		w.Write([]byte(string("No public key given")))
		return
	}
	publicKey := strings.TrimSpace(userUpload.PublicKey)
	log.Println("Pub Key: ", publicKey)
	sum := sha3.Sum256([]byte(publicKey))
	hashedKey := hex.EncodeToString(sum[:])
	//var transactions []data2.Transaction
	transactions := ""
	for key := range SBC.GetBlockChain().Chain {
		value := SBC.GetBlockChain().Chain[key]
		for index := range value {
			blockInChain := value[index]
			transactionJson, _ := blockInChain.Value.Get(hashedKey)
			if transactionJson != "" && transactionJson != "failure" { //transaction exists in this mpt for Public Key
				//var transaction data2.Transaction
				//err := json.Unmarshal([]byte(transactionJson), &transaction)
				//if err != nil {
				//	log.Println(err, "Error in converting JSON to transaction object ")
				//	w.WriteHeader(500)
				//	return
				//}
				//transactions = append(transactions, transaction)
				transactions += transactionJson + "\n\n"
			}
		}
	}
	if transactions == "" {
		w.WriteHeader(200)
		w.Write([]byte(string("No record exist with Public Key in Blockchain")))
	} else {
		//transactionsJson, _ := json.Marshal(transactions)
		w.WriteHeader(200)
		w.Write([]byte(string(transactions)))
	}
}

//if transaction queue isn't empty, take next transaction in front of the queue
func removeTransactionFromPool() data2.Transaction {
	return data2.Transaction{}
}

func FindReviewsByProductAndUserID(w http.ResponseWriter, r *http.Request) {
	log.Println("Entered FindReviewsByProductID")
	var userUpload data2.UserWebUpload
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err, "Error in reading post request: FindReviewsByProductID ")
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal([]byte(request), &userUpload)
	if err != nil {
		log.Println(err, "Error in json unmarshal: FindReviewsByProductID ")
		w.WriteHeader(500)
		return
	}
	if userUpload.PublicKey == "" {
		log.Println(err, "No public key given")
		w.WriteHeader(500)
		w.Write([]byte(string("No public key given")))
		return
	}
	if userUpload.ProductID == "" {
		log.Println(err, "No product id given")
		w.WriteHeader(500)
		w.Write([]byte(string("No product id given")))
		return
	}
	publicKey := strings.TrimSpace(userUpload.PublicKey)
	log.Println("Pub Key: ", publicKey)
	sum := sha3.Sum256([]byte(publicKey))
	hashedKey := hex.EncodeToString(sum[:])
	//var transactions []data2.Transaction
	transactions := ""
	for key := range SBC.GetBlockChain().Chain {
		value := SBC.GetBlockChain().Chain[key]
		for index := range value {
			blockInChain := value[index]
			transactionJson, _ := blockInChain.Value.Get(hashedKey)
			if transactionJson != "" && transactionJson != "failure" { //transaction exists in this mpt for Public Key
				var transaction data2.Transaction
				err := json.Unmarshal([]byte(transactionJson), &transaction)
				if err != nil {
					log.Println(err, "Error in converting JSON to transaction object ")
					w.WriteHeader(500)
					return
				}
				//transactions = append(transactions, transaction)
				if transaction.ReviewObj.Product.ProductID == userUpload.ProductID {
					transactions += transactionJson + "\n\n"
				}
			}
		}
	}
	if transactions == "" {
		w.WriteHeader(200)
		w.Write([]byte(string("No record exist for Public Key/ProductID pair in Blockchain")))
	} else {
		//transactionsJson, _ := json.Marshal(transactions)
		w.WriteHeader(200)
		w.Write([]byte(string(transactions)))
	}
}
