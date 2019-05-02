package p5

import (
	"bytes"
	"cs686-blockchain-p3-mosopeogundipe/p5/data"
	"cs686-blockchain-p3-mosopeogundipe/p5/logic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//Port 7700 is used for this server

var PRODUCT data.Product
var PRODUCTS data.Products
var USERS data.Users
var REVIEW data.Review
var MINERS map[string]int32 //Stores (K,V) - > ("http://IP:Port", "Port")
var USERUPLOAD data.UserWebUpload
var MINERS_TRANSACTION_API = "/transaction/receive"

func init() {
}

// API #1 in user flow
// Creates user id and public private key, stores others and returns only private key to user
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	publicKey, privateKey := logic.CreateUser(USERS)
	var value string = ""
	value += publicKey + "\n"
	value += privateKey
	fmt.Fprintf(w, value)
}

//API #1 in merchant flow
//Registers a product to the database of products. User provides product name and ID in JSON like:
/*{
"ProductName": "Smart Water",
"ProductID": "A0372926671"
}*/
//Product ID is GTIN of product (e.g. the barcode number)
func RegisterProduct(w http.ResponseWriter, r *http.Request) {
	log.Println("In RegisterProduct")
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err, "Error in reading post request: RegisterProduct ")
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal([]byte(request), &PRODUCT)
	if err != nil {
		log.Println(err, "Error in json unmarshal: RegisterProduct ")
		w.WriteHeader(500)
		return
	}
	successful, alreadyexists := logic.AddProduct(PRODUCT, PRODUCTS)
	if successful {
		w.WriteHeader(200)
		w.Write([]byte("Product registration successful"))
	} else if !successful && alreadyexists == "exists" {
		w.WriteHeader(200)
		w.Write([]byte("There's already a product for ID " + PRODUCT.ProductID))
	}

}

// API #1 in miner flow
// Miners call this API to register themselves in this central data layer
func RegisterMiner(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	host := q["host"][0]
	portStr := q["id"][0]
	addr := "http://" + host
	log.Println("In RegisterMiner host: ", host, "Port: ", portStr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err, "RegisterMiner")
		w.WriteHeader(400)
		return
	}
	if len(MINERS) == 0 {
		MINERS = make(map[string]int32)
	}
	if MINERS[addr] > 0 {
		log.Println("Miner " + addr + "Already registered")
		w.WriteHeader(200)
	} else {
		MINERS[addr] = int32(port)
		w.WriteHeader(200)
	}
}

// API #1 in user flow
// User uploads a review here
// This API accepts json that looks like:
/*{
"PublicKey": "xhsxhsgxshkxgshdsygsjkxxvvssjkxjhsxjxvbxvcv",
"ProductID": "A0372926671",
"Review": "this was a horrible product"
}*/
func PostReview(w http.ResponseWriter, r *http.Request) {
	log.Println("In PostReview")
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err, "Error in reading post request: PostReview ")
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal([]byte(request), &USERUPLOAD)
	if err != nil {
		log.Println(err, "Error in json unmarshal: PostReview ")
		w.WriteHeader(500)
		return
	}
	if USERUPLOAD.Signature == "" {
		log.Println(err, "Invalid Message: No Signature Sent ")
		w.WriteHeader(200)
		w.Write([]byte("Invalid Message: No Signature Sent"))
		return
	}
	product, productfound := PRODUCTS.ProductSet[USERUPLOAD.ProductID]
	sum := sha3.Sum256([]byte(USERUPLOAD.PublicKey))
	_, userfound := USERS.UserSet[hex.EncodeToString(sum[:])]
	if productfound == false {
		log.Println("Product with ID" + USERUPLOAD.ProductID + "does not exist in database ")
		w.WriteHeader(404)
		return
	}
	if userfound == false {
		log.Println("Product does not exist in database ")
		w.WriteHeader(404)
		return
	}
	review := data.Review{Product: product, Review: USERUPLOAD.Review}
	transaction := data.Transaction{TransactionID: "", PublicKey: USERUPLOAD.PublicKey, ReviewObj: review}
	transactionJson, _ := json.Marshal(transaction)
	sum = sha3.Sum256([]byte(transactionJson))
	transactionHashNoID := hex.EncodeToString(sum[:])
	transaction.TransactionID = transactionHashNoID
	transactionJson, _ = json.Marshal(transaction) //send this json to all miners using POST API
	ForwardTransactionToMiners(string(transactionJson), USERUPLOAD.Signature)

}

func ForwardTransactionToMiners(transactionJson string, signature string) {
	for addr, _ := range MINERS {
		url := addr + MINERS_TRANSACTION_API + "?Signature=" + signature
		httpPost(url, transactionJson)
	}
}

//API called by to get signature of a message for a user
//This simulates the action of user signing a message with their private key
//It returns the signature of the message
func SignMessage(w http.ResponseWriter, r *http.Request) {
	var signMessage data.SignMessage
	log.Println("In SignMessage")
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err, "Error in reading post request: SignMessage ")
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal([]byte(request), &signMessage)
	if err != nil {
		log.Println(err, "Error in json unmarshal: SignMessage ")
		w.WriteHeader(500)
		return
	}
	product := PRODUCTS.ProductSet[signMessage.ProductID]
	review := data.Review{Product: product, Review: signMessage.Review}
	transaction := data.Transaction{TransactionID: "", PublicKey: "", ReviewObj: review}
	transactionJson, _ := json.Marshal(transaction)
	signature := logic.SignWithPrivateKey([]byte(transactionJson), []byte(signMessage.PrivateKey))
	w.WriteHeader(200)
	w.Write(signature)
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
