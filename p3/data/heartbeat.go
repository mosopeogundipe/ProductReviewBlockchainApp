package data

import (
	"cs686-blockchain-p3-mosopeogundipe/p1"
	"cs686-blockchain-p3-mosopeogundipe/p2"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	var heartBeat HeartBeatData
	if !ifNewBlock {
		heartBeat = HeartBeatData{IfNewBlock: ifNewBlock, Id: id, BlockJson: "", PeerMapJson: peerMapJson, Addr: addr, Hops: 2}
	} else {
		heartBeat = HeartBeatData{IfNewBlock: ifNewBlock, Id: id, BlockJson: blockJson, PeerMapJson: peerMapJson, Addr: addr, Hops: 2}
	}
	return heartBeat
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string) HeartBeatData {
	if willCreateNewBlock() {
		mptStringMap := make(map[string]string)
		length := rand.Intn(10)
		memberLength := 5
		for i := 0; i < length; i++ {
			key := generateRandomString(memberLength)
			value := generateRandomString(memberLength)
			mptStringMap[key] = value
		}
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()
		for key, value := range mptStringMap {
			mpt.Insert(key, value)
		}
		_, isGenesisBlockCreated := sbc.Get(0)
		var block p2.Block
		var parent []p2.Block
		if !isGenesisBlockCreated {
			block = sbc.GenBlock(mpt)
		} else {
			parent, _ = sbc.Get(sbc.bc.Length)
			block = p2.Initial(parent[0].Header.Hash, parent[0].Header.Height, mpt)
		}
		sbc.Insert(block)
		return NewHeartBeatData(true, selfId, block.EncodeToJSON(), peerMapJSON, addr)
	} else {
		return NewHeartBeatData(false, selfId, "", peerMapJSON, addr)
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

func generateRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
