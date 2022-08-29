package ls2gs

import (
	"l2goserver/loginserver/types/dao"
	"l2goserver/packets"
)

func AuthedResponse(serverId byte) *packets.Buffer {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x02)
	buffer.WriteSingleByte(serverId)
	buffer.WriteS(dao.GetServerNameById(serverId))
	return buffer
}
