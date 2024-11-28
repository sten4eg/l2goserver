package ls2gs

import (
	"l2goserver/loginserver/types/dao"
	"l2goserver/packets"
)

func AuthedResponse(serverId byte) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x02)
	buffer.WriteSingleByte(serverId)
	buffer.WriteS(dao.GetServerNameById(serverId))
	return buffer.CopyBytes()
}
