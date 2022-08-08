package clientpackets

import (
	"errors"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
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

	buff := packets.Get()

	if !(key1 == client.SessionKey.LoginOk1 && key2 == client.SessionKey.LoginOk2) {
		err = client.SendBuf(serverpackets.NewLoginFailPacket(reason.AccessFailed, buff))
		if err != nil {
			return err
		}
		return serverOverload
	}

	//TODO коннект к гейм серверу и проверка можно ли к нему подконектиться
	if true {
		client.SendBuf(serverpackets.NewPlayOkPacket(client, buff))
		client.JoinedGS = true
	} else {
		client.SendBuf(serverpackets.NewPlayFailPacket(reason.ServerOverloaded, buff))
		return serverOverload
	}

	return nil
}
