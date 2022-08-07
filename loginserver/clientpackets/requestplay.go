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

	if !(key1 == client.SessionKey.LoginOk1 && key2 == client.SessionKey.LoginOk2) {
		err = client.Send(serverpackets.NewLoginFailPacket(reason.AccessFailed))
		if err != nil {
			return err
		}
	}

	//TODO коннект к гейм серверу и проверка можно ли к нему подконектиться
	if true {
		client.Send(serverpackets.NewPlayOkPacket(client))
		client.JoinedGS = true
	} else {
		client.Send(serverpackets.NewPlayFailPacket(reason.ServerOverloaded))
		return serverOverload
	}

	return nil
}
