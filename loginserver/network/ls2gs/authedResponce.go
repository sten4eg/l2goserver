package ls2gs

import (
	"l2goserver/loginserver/types/dao"
	"l2goserver/packets"
)

func AuthedResponse(serverId byte) *packets.Buffer {
	buf := packets.Get()
	buf.WriteSingleByte(0x02)
	buf.WriteSingleByte(serverId)
	buf.WriteS(dao.GetServerNameById(serverId))
	return buf
}
