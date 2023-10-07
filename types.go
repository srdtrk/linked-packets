package linkedpackets

import "encoding/json"

const LastLinkMemoKey = "last_link"

type LinkData struct {
	LinkID         string           `json:"link_id"`
	PrevPacket     PacketIdentifier `json:"prev_packet"`
	IsLastPacket   bool             `json:"last_packet"`
	IsInitalPacket bool             `json:"initial_packet"`
	LinkIndex      string           `json:"link_index"`
}

func (ld LinkData) String() string {
	bz, err := json.Marshal(ld)
	if err != nil {
		panic(err)
	}
	return string(bz)
}
