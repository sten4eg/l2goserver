package dao

var serverName = []string{
	"Undefined",
	"Bartz",
	"Sieghardt",
	"Kain",
	"Lionna",
	"Erica",
	"Gustin",
} //todo продолжение

func GetServerNameById(id byte) string {
	idI := int(id)
	if len(serverName) > idI {
		return serverName[idI]
	}
	return serverName[0]
}
