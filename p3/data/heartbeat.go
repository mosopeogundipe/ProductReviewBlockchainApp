package data

import (
	"cs686-blockchain-p3-mosopeogundipe/p1"
	"cs686-blockchain-p3-mosopeogundipe/p2"
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"log"
	"math/rand"
	"strings"
)

var LETTER_RUNES = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var HEX_CHARSET = []rune("abcdefABCDEF123456789")
var IS_PARENT_USED_ALREADY = false
var NONCE_PREFIX = "00000"

//var isNewBlck = false
//var block p2.Block

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

//func SetValues(isNewBlock bool, theBlock p2.Block){
//	isNewBlck = isNewBlock
//	block = theBlock
//}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	var heartBeat HeartBeatData
	if !ifNewBlock {
		heartBeat = HeartBeatData{IfNewBlock: ifNewBlock, Id: id, BlockJson: "", PeerMapJson: peerMapJson, Addr: addr, Hops: 2}
	} else {
		heartBeat = HeartBeatData{IfNewBlock: ifNewBlock, Id: id, BlockJson: blockJson, PeerMapJson: peerMapJson, Addr: addr, Hops: 2}
	}
	return heartBeat
}

//OLD implementation - randomly generated new block
//func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string) HeartBeatData {
//	if willCreateNewBlock() {
//		mptStringMap := make(map[string]string)
//		length := rand.Intn(10)
//		memberLength := 5
//		for i := 0; i < length; i++ {
//			key := generateRandomString(memberLength)
//			value := generateRandomString(memberLength)
//			mptStringMap[key] = value
//		}
//		mpt := p1.MerklePatriciaTrie{}
//		mpt.Initial()
//		for key, value := range mptStringMap {
//			mpt.Insert(key, value)
//		}
//		_, isGenesisBlockCreated := sbc.Get(0)
//		var block p2.Block
//		var parent []p2.Block
//		if !isGenesisBlockCreated {
//			block = sbc.GenBlock(mpt)
//		} else {
//			parent, _ = sbc.Get(sbc.bc.Length)
//			block = p2.Initial(parent[0].Header.Hash, parent[0].Header.Height, mpt)
//		}
//		sbc.Insert(block)
//		return NewHeartBeatData(true, selfId, block.EncodeToJSON(), peerMapJSON, addr)
//	} else {
//		return NewHeartBeatData(false, selfId, "", peerMapJSON, addr)
//	}
//}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string, block p2.Block, isNewBlck bool) HeartBeatData {
	if isNewBlck {
		sbc.Insert(block)
		return NewHeartBeatData(isNewBlck, selfId, block.EncodeToJSON(), peerMapJSON, addr)
	} else {
		return NewHeartBeatData(isNewBlck, selfId, "", peerMapJSON, addr)
	}
}

func willCreateNewBlock() bool {
	random := rand.Intn(100)
	if random < 50 {
		return false
	} else {
		return true
	}
}

func CreateRandomMpt() p1.MerklePatriciaTrie {
	mptStringMap := make(map[string]string)
	length := rand.Intn(10)
	memberLength := 5
	for i := 0; i < length; i++ {
		key := generateRandomString(memberLength, LETTER_RUNES)
		value := generateRandomString(memberLength, LETTER_RUNES)
		mptStringMap[key] = value
	}
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	for key, value := range mptStringMap {
		mpt.Insert(key, value)
	}
	return mpt
}

func GetNonce(parentBlock p2.Block, mptRoot string) string {
	var isNonceGenerated bool = false
	var nonce string = ""
	parentHash := parentBlock.Header.Hash
	for !isNonceGenerated && !IS_PARENT_USED_ALREADY {
		//for parentBlock.Header.Height == SBC.GetLatestBlocks()[0].Header.Height{
		nonce = generateRandomString(16, HEX_CHARSET)
		value := sha3.Sum256([]byte(parentHash + nonce + mptRoot))
		valueStr := hex.EncodeToString(value[:]) // Do we need to convert to string to verify 0s OR should we just check 'value' byte array for successive 0s?
		if strings.HasPrefix(valueStr, NONCE_PREFIX) {
			log.Println("Nonce found! Value verified: ", valueStr)
			isNonceGenerated = true
			break
		} else {
			continue
		}
	}
	return nonce
}

func generateRandomString(n int, charset []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
