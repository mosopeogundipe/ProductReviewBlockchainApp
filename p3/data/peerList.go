package data

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32 //Stores (K,V) - > ("http://IP:Port", "Id")
	maxLength int32
	mux       sync.Mutex
}

type PeerMap struct {
	Addr string `json:"addr"`
	Id   int32  `json:"id"`
}

// A data structure to hold key/value pairs
type Pair struct {
	Key   string
	Value int32
}

// A slice of pairs that implements sort.Interface to sort by values
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

func NewPeerList(id int32, maxLength int32) PeerList {
	return PeerList{selfId: id, peerMap: make(map[string]int32), maxLength: maxLength, mux: sync.Mutex{}}
}

func (peers *PeerList) Add(addr string, id int32) {
	log.Println("Enters Peer Add method")
	//add into peermap only if it isn't already there!
	peers.mux.Lock()
	if peers.peerMap[addr] != id {
		log.Println("Adds to peermap|| Addr: ", addr, "id: ", id)
		peers.peerMap[addr] = id
	}
	peers.mux.Unlock()
	log.Println("Leaving Peer Add method")
}

func (peers *PeerList) Delete(addr string) {
	delete(peers.peerMap, addr)
}

func (peers *PeerList) Rebalance() {
	if len(peers.peerMap) > 0 {
		p := make(PairList, len(peers.peerMap))
		i := 0
		for k, v := range peers.peerMap {
			p[i] = Pair{k, v}
			i++
		}
		sort.Sort(p)
		peers.peerMap = peers.createPeerMap(p)
	} else {
		log.Println("No need to rebalance, peermap contains no element")
	}
}

func (peers *PeerList) createPeerMap(listOfPairs PairList) map[string]int32 {
	//log.Println("self id: ", peers.selfId)
	//fmt.Println("list of pairs", listOfPairs)
	counter := int32(peers.maxLength / 2)
	leftIndex := 0
	rightIndex := 0
	retVal := make(map[string]int32)
	for i := 0; i < len(listOfPairs)-1; i++ {
		if listOfPairs[i].Value < peers.selfId && listOfPairs[i+1].Value > peers.selfId {
			leftIndex = i
			rightIndex = i + 1
			break
		}
	}
	if leftIndex == 0 && rightIndex == 0 {
		return peers.peerMap
	}
	fmt.Println("left Index", leftIndex)
	fmt.Println("right Index", rightIndex)
	for counter > 0 { //get the left half of rebalanced map
		index := leftIndex
		//fmt.Println("index: ", index)
		if leftIndex < 0 {
			index = len(listOfPairs) + leftIndex
			fmt.Println("index in left part: ", index)
			if index > 0 { //ensure only valid indices are accessed
				retVal[listOfPairs[index].Key] = listOfPairs[index].Value
			}
		} else {
			retVal[listOfPairs[index].Key] = listOfPairs[index].Value
		}
		leftIndex--
		counter--
	}
	counter = int32(peers.maxLength / 2)
	for counter > 0 { //get the left half of rebalanced map
		index := rightIndex
		if rightIndex > len(listOfPairs)-1 {
			index = rightIndex - len(listOfPairs)
			fmt.Println("index in right part: ", index)
			if index < len(listOfPairs)-1 {
				retVal[listOfPairs[index].Key] = listOfPairs[index].Value
			}
		} else {
			retVal[listOfPairs[index].Key] = listOfPairs[index].Value
		}
		rightIndex++
		counter--
	}
	return retVal
}

func (peers *PeerList) Show() string {
	var result = "This is PeerMap:" + "\n"
	for k, v := range peers.peerMap {
		result += "addr: " + k + "id: " + string(v)
	}
	return result
}

func (peers *PeerList) Register(id int32) {
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
}

func (peers *PeerList) Copy() map[string]int32 {
	temp := peers.peerMap
	return temp
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	result, err := json.Marshal(peers.peerMap)
	return string(result), err
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	var peerMap PeerMap
	err := json.Unmarshal([]byte(peerMapJsonStr), &peerMap)
	if err != nil {
		log.Fatal("Error in InjectPeerMapJson: ", err)
	}
	//for i:=0; i<len(peerMapList); i++{
	//
	//}
	if peerMap.Addr != selfAddr {
		log.Println("adding to peerlist. Addr: ", peerMap.Addr, "port: ", peerMap.Id)
		//peers.peerMap[peerMapList[i].Addr] = peerMapList[i].Id
		if strings.Contains(peerMap.Addr, "http://") {
			peers.Add(peerMap.Addr, peerMap.Id)
		} else {
			peers.Add("http://"+peerMap.Addr, peerMap.Id)
		}
	}
	log.Println("Leaving InjectPeerMapJson")
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println("actual: ", peers)
	fmt.Println("expected: ", expected)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}

//func main() {
//	TestPeerListRebalance()
//}
