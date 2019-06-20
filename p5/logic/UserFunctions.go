package logic

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"cs686-blockchain-p3-mosopeogundipe/p5/data"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
	"os"
	"strings"
)

//Stores hash of user's public key in set
func CreateUser(users *data.Users) (string, string) {
	publicKey, privateKey := createPublicPrivateKeyPair()
	publicKeyStr := ExportRsaPublicKeyAsPemStr(publicKey)
	privateKeyStr := ExportRsaPrivateKeyAsPemStr(privateKey)
	publicKeyStr = strings.TrimSpace(publicKeyStr)
	sum := sha3.Sum256([]byte(publicKeyStr))
	hash := hex.EncodeToString(sum[:])
	log.Println("Public Key: ", publicKeyStr)
	if len(users.UserSet) == 0 {
		users.UserSet = make(map[string]bool)
		users.UserSet[hash] = true
		//log.Println("storing in user set. len: ", len(users.UserSet))
	} else {
		users.UserSet[hash] = true
	}
	log.Println("User PK Hash: ", hash)
	return publicKeyStr, privateKeyStr
}

//creates key using RSA
func createPublicPrivateKeyPair() (*rsa.PublicKey, *rsa.PrivateKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("error in creating RSA Keys", err.Error)
		os.Exit(1)
	}
	publicKey := &privateKey.PublicKey
	return publicKey, privateKey
}

//converts private key to string to be given to user
func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

//converts public key to string to be given to user
func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) string {
	pubkey_bytes := x509.MarshalPKCS1PublicKey(pubkey)
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem)
}

//signs a message with user's private key
func SignWithPrivateKey(message []byte, privatekey []byte) []byte {
	PKCS1message := message
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PKCS1message)
	hashed := pssh.Sum(nil)
	log.Println("Priv Key Sign, hash message: ", hashed)
	block, _ := pem.Decode(privatekey)
	if block == nil {
		fmt.Fprintf(os.Stderr, "Error in private key %s\n")
		return []byte{}
	}
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing private key from bytes: %s\n", err)
		return []byte{}
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, newhash, hashed[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return signature
	}
	sig := base64.StdEncoding.EncodeToString(signature) //the raw "signature" byte array uses a weird encoding and gives foreign characters that can't be sent over Web API
	log.Println("Signature: ", sig)
	return []byte(sig)
}

//verifies that message is signed by correct user
func VerifyPrivateKeySignature(message []byte, signature []byte, publicKey []byte) bool {
	PKCS1message := message
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PKCS1message)
	hashed := pssh.Sum(nil)
	log.Println("Pub Key Sign, hash message: ", hashed)
	block, _ := pem.Decode(publicKey)
	if block == nil {
		fmt.Fprintf(os.Stderr, "Error in public key %s\n")
		return false
	}
	//pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	rsaPublicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing public key from bytes: %s\n", err)
		return false
	}
	//rsaPublicKey := pubInterface.(*rsa.PublicKey)
	var errs error = rsa.VerifyPKCS1v15(rsaPublicKey, newhash, hashed[:], signature)
	if errs != nil {
		fmt.Fprintf(os.Stderr, "Error from signature verification: %s\n", errs)
		return false
	}
	return true
}
