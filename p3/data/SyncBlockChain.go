package data

import (
	"cs686-blockchain-p3-mosopeogundipe/p1"
	"cs686-blockchain-p3-mosopeogundipe/p2"
	"errors"
	"log"
	"sync"
)

var queue []p2.Block

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

//HINT: PREVIOUS INSERT FUNCTION COMMENTED OUT -- Enhanced performance with new logic below
//func (sbc *SyncBlockChain) Insert(block p2.Block) {
//	sbc.mux.Lock()
//	sbc.bc.Insert(block)
//	sbc.mux.Unlock()
//}

//In order to reduce network losses, I changed insert method to add to queue and have another thread to perform inserts later.
//This helps ensures that forks in later chains get stored in all nodes
func (sbc *SyncBlockChain) Insert(block p2.Block) { //simply adds the block to a queue here
	sbc.mux.Lock()
	//log.Println("INSERT : appending to queue")
	queue = append(queue, block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) FinishInsert() { //inserts all available elements of queue into blockchain
	log.Println("Entered FINISH INSERT")
	for {
		sbc.mux.Lock()
		if len(queue) > 0 {
			sbc.bc.Insert(queue[0])
			// Dequeue
			queue[0] = p2.Block{} // Erase element (write empty block)
			queue = queue[1:]
		}
		sbc.mux.Unlock()
	}
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	//return sbc.bc.IsParentBlockInBlockChain(insertBlock.Header.ParentHash)
	return sbc.bc.IsParentBlockInBlockChain(insertBlock)
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

func (sbc *SyncBlockChain) GetBlockChain() p2.BlockChain {
	return sbc.bc
}
