package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	regData := RegisterData{AssignedId: id, PeerMapJson: peerMapJson}
	return regData
}

func (data *RegisterData) EncodeToJson() (string, error) {
	result, err := json.Marshal(data)
	return string(result), err
}
