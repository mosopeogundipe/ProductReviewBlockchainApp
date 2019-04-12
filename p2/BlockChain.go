package p2

import (
	"cs686-blockchain-p3-mosopeogundipe/p1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"sort"
	"strings"
	"time"
)

type Header struct {
	Height     int32
	Timestamp  int64
	Hash       string
	ParentHash string
	Size       int32
	Nonce      string
}

type Block struct {
	Header Header
	Value  p1.MerklePatriciaTrie
}

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

type BlockJson struct {
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timeStamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	Nonce      string            `json:"nonce"`
	MPT        map[string]string `json:"mpt"`
}

func NewBlockChain() BlockChain {
	return BlockChain{Chain: make(map[int32][]Block), Length: 0}
}

/**
Creates a block based on the MerklePatriciaTree, parent's height and parent hash
*/
func Initial(parent_hash string, parent_height int32, value p1.MerklePatriciaTrie) Block {
	//If parent hash is empty and no block exists in blockchain, create genesis block
	if parent_hash == "" {
		return CreateGenesisBlock(value)
	}
	height := parent_height + 1
	time_stamp := time.Now().UnixNano() / 1000000 //Current time in Unix milliseconds
	size := int32(len([]byte(value.String())))    //confirm if this approach is fine
	header := Header{Height: height, Timestamp: time_stamp, Hash: "", ParentHash: parent_hash, Size: size}
	block := Block{Header: header, Value: value}
	block.Header.Hash = block.hash_block(value)
	return block
}

/**
Hashes a block instance
*/
func (b *Block) hash_block(value p1.MerklePatriciaTrie) string {
	//hash_str := string(header.Height) + string(header.Timestamp) + header.ParentHash + value.GetRoot() + string(header.Size)
	hash_str := string(b.Header.Height) + string(b.Header.Timestamp) + b.Header.ParentHash + value.GetRoot() + string(b.Header.Size)
	hash := sha3.Sum256([]byte(hash_str))
	hash_str = hex.EncodeToString(hash[:])
	return hash_str
}

/**
Creates a Genesis Block - the first block in the blockchain if blockchain is empty
*/
func CreateGenesisBlock(value p1.MerklePatriciaTrie) Block {
	height := int32(0)
	time_stamp := time.Now().UnixNano() / 1000000
	size := int32(len([]byte(value.String())))
	header := Header{Height: height, Timestamp: time_stamp, Hash: "", ParentHash: "genesis", Size: size}
	block := Block{Header: header, Value: value}
	block.Header.Hash = block.hash_block(value)
	return block
}

/**
Serializes a block to JSON String
*/
func (b *Block) EncodeToJSON() string {
	//	fmt.Println("BLOCK KEY VAL: ", b.Value.GetMptKeyValues())
	blockJsonObject := BlockJson{Height: b.Header.Height, Timestamp: b.Header.Timestamp,
		Hash: b.Header.Hash, ParentHash: b.Header.ParentHash, Size: b.Header.Size, Nonce: b.Header.Nonce, MPT: b.Value.KeyVal}
	blockJsonStr, _ := json.Marshal(blockJsonObject)
	return string(blockJsonStr)
}

/**
Deserializes a block from JSON string to Block Object
*/
func (b *Block) DecodeFromJson(jsonString string) Block {
	blockJsonObject := BlockJson{}
	json.Unmarshal([]byte(jsonString), &blockJsonObject)
	return convertBlockJsonToBlock(blockJsonObject)
}

/**
Converts from a BlockJson structure to a Block Structure
*/
func convertBlockJsonToBlock(blockJsonObject BlockJson) Block {
	blockHeader := Header{Height: blockJsonObject.Height, Timestamp: blockJsonObject.Timestamp, Hash: blockJsonObject.Hash,
		ParentHash: blockJsonObject.ParentHash, Size: blockJsonObject.Size, Nonce: blockJsonObject.Nonce}
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	//fmt.Println("MPT: ", blockJsonObject.MPT)
	for key, value := range blockJsonObject.MPT {
		mpt.Insert(key, value)
	}
	block := Block{Header: blockHeader, Value: mpt}
	//fmt.Println(block)
	return block
}

/**
Serializes a Blockchain structure to a JSON string
*/
func (bc *BlockChain) EncodeToJSON() string {
	var result []string
	var blockJson string
	//fmt.Println("BC LEN: ", len(bc.Chain))
	for key := range bc.Chain {
		value := bc.Chain[key]
		fmt.Println("VALUE:", value)
		//fmt.Println("LIST LEN: ", len(value))
		for index := range value {
			blockJson = value[index].EncodeToJSON()
			result = append(result, blockJson)
		}
	}
	//fmt.Println("RESULTING BC JSON: ", result)
	return "[" + strings.Join(result, ",") + "]"
}

/**
Deserializes a JSON string to a Blockchain structure
*/
func (bc *BlockChain) DecodeFromJSON(jsonString string) BlockChain {
	var blockJsonList []BlockJson
	json.Unmarshal([]byte(jsonString), &blockJsonList)
	var blockChain BlockChain
	for index := range blockJsonList {
		blockJson := blockJsonList[index]
		block := convertBlockJsonToBlock(blockJson)
		blockChain.Insert(block)
	}
	return blockChain
}

/**
Returns List of Blocks at a soecified height
*/
func (bc *BlockChain) Get(height int32) []Block {
	if bc.Chain[height] == nil || len(bc.Chain[height]) == 0 {
		return nil
	} else {
		return bc.Chain[height]
	}
}

/**
Inserts a Block into the Blockchain
*/
func (bc *BlockChain) Insert(block Block) {
	//if !bc.isBlockInBlockChain(block) {
	currentBlockList := bc.Get(block.Header.Height)
	if len(currentBlockList) == 0 {
		fmt.Println("blockahain is empty at height: ", block.Header.Height)
		newChain := []Block{}
		newChain = append(newChain, block)
		if len(bc.Chain) == 0 {
			bc.Chain = make(map[int32][]Block)
		}
		bc.Chain[block.Header.Height] = newChain
		bc.Length = block.Header.Height
	} else {
		fmt.Println("blockchain is not empty at height: ", block.Header.Height)
		for _, currBlock := range currentBlockList {
			if block.Header.Hash == currBlock.Header.Hash {
				return
			}
		}
		bc.Chain[block.Header.Height] = append(bc.Chain[block.Header.Height], block)
	}
	maxHeight := bc.FindMaxHeight()
	bc.Length = maxHeight
	//} else {
	//fmt.Println("block with hash ", block.Header.Hash, "already exists in blockchain")
	//}
}

/**
Returns the maximum height of the BlockChain
*/
func (bc *BlockChain) FindMaxHeight() int32 {
	var maxIndex int32 = 0
	for index := range bc.Chain {
		if index > maxIndex {
			maxIndex = index
		}
	}
	return maxIndex
}

func (bc *BlockChain) isBlockInBlockChain(block Block) bool {
	for key := range bc.Chain {
		value := bc.Chain[key]
		for index := range value {
			blockInChain := value[index]
			if blockInChain.Header.Hash == block.Header.Hash {
				return true
			}
		}
	}
	return false
}

func (bc *BlockChain) IsParentBlockInBlockChain(block Block) bool {
	blockList := bc.Get(block.Header.Height - 1)
	for _, currBlock := range blockList {
		if block.Header.ParentHash == currBlock.Header.Hash {
			return true
		}
	}
	return false
}

func (bc *BlockChain) GetLatestBlocks() []Block {
	return bc.Get(bc.Length)
}

func (bc *BlockChain) GetParentBlock(block Block) Block {
	for key := range bc.Chain {
		value := bc.Chain[key]
		for index := range value {
			blockInChain := value[index]
			if blockInChain.Header.Hash == block.Header.ParentHash {
				return blockInChain
			}
		}
	}
	return Block{}
}

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}
