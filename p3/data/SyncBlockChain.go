package data

import (
	"cs686-blockchain-p3-mosopeogundipe/p1"
	"cs686-blockchain-p3-mosopeogundipe/p2"
	"errors"
	"sync"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	var blocksAtHeight = sbc.bc.Get(height)
	var boolean = false
	if len(blocksAtHeight) > 0 {
		boolean = true
	}
	return blocksAtHeight, boolean
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	var blockList []p2.Block = sbc.bc.Get(height)
	for i := 0; i < len(blockList); i++ {
		if blockList[i].Header.Hash == hash {
			return blockList[i], true
		}
	}
	return p2.Block{}, false
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.IsParentBlockInBlockChain(insertBlock.Header.ParentHash)
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	sbc.bc = sbc.bc.DecodeFromJSON(blockChainJson)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	result := sbc.bc.EncodeToJSON()
	if len(sbc.bc.Chain) == 0 {
		return result, errors.New("blockchain is empty")
	} else {
		return result, nil
	}
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	genesisBlock := p2.CreateGenesisBlock(mpt)
	return genesisBlock
}

func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	latestBlocks := sbc.bc.GetLatestBlocks()
	return latestBlocks
}

func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	parentBlock := sbc.bc.GetParentBlock(block)
	return parentBlock
}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}
