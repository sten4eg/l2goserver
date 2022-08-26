package c2ls

import (
	"errors"
	"l2goserver/loginserver/models"
	serverpackets2 "l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/reason"
	"l2goserver/packets"
)

var errServerOverload = errors.New("serverOverload")

func RequestServerLogin(request []byte, client *models.ClientCtx) error {
	var packet = packets.NewReader(request)
	var err error

	key1 := packet.ReadUInt32()
	key2 := packet.ReadUInt32()
	serverId := packet.ReadUInt8()

	_ = serverId

	if key1 != client.SessionKey.LoginOk1 || key2 != client.SessionKey.LoginOk2 {
		err = client.SendBuf(serverpackets2.NewLoginFailPacket(reason.AccessFailed))
		if err != nil {
			return err
		}
		return errServerOverload
	}

	//TODO коннект к гейм серверу и проверка можно ли к нему подконектиться
	if true {
		client.SetJoinedGS(true)
		err = client.SendBuf(serverpackets2.NewPlayOkPacket(client))
	} else {
		_ = client.SendBuf(serverpackets2.NewPlayFailPacket(reason.ServerOverloaded))
		return errServerOverload
	}

	return err
}
