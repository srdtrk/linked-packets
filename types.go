package linkedpackets

type LinkData struct {
	LinkID string `json:"link_id"`
	PrevPacket PacketIdentifier `json:"prev_packet"`
	IsLastPacket bool `json:"last_packet"`
	IsInitalPacket bool `json:"initial_packet"`
}

type PacketIdentifier struct {
	Sequence string `json:"seq"`
	ChannelID string `json:"channel_id"`
	PortID string `json:"port_id"`
}
