package ls2gs

import (
	"l2goserver/loginserver/types/serverNames"
	"l2goserver/packets"
)

func AuthedResponse(serverId byte) *packets.Buffer {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x02)
	buffer.WriteSingleByte(serverId)
	buffer.WriteS(serverNames.GetServerNameById(serverId))
	return buffer
}
