package serverNames

var serverName = []string{
	"Undefined",
	"Bartzz",
	"Sieghardt",
	"Kain",
	"Lionna",
	"Erica",
	"Gustin",
}

func GetServerNameById(id byte) string {
	idI := int(id)
	if len(serverName) > idI {
		return serverName[idI]
	}
	return serverName[0]
}
