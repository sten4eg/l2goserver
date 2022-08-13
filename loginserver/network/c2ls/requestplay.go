package c2ls

import (
	"errors"
	"l2goserver/loginserver/models"
	serverpackets2 "l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/reason"
	"l2goserver/packets"
)

var serverOverload = errors.New("serverOverload")

func NewRequestPlay(request []byte, client *models.ClientCtx) error {
	var packet = packets.NewReader(request)
	var err error

	key1 := packet.ReadUInt32()
	key2 := packet.ReadUInt32()
	serverId := packet.ReadUInt8()

	_ = serverId
	_ = key2
	_ = key1
	_ = err

	if !(key1 == client.SessionKey.LoginOk1 && key2 == client.SessionKey.LoginOk2) {
		err = client.SendBuf(serverpackets2.NewLoginFailPacket(reason.AccessFailed))
		if err != nil {
			return err
		}
		return serverOverload
	}

	//TODO коннект к гейм серверу и проверка можно ли к нему подконектиться
	if true {
		err = client.SendBuf(serverpackets2.NewPlayOkPacket(client))
		client.JoinedGS = true
	} else {
		_ = client.SendBuf(serverpackets2.NewPlayFailPacket(reason.ServerOverloaded))
		return serverOverload
	}

	return err
}
