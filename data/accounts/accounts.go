package accounts

import (
	"github.com/jackc/pgx"
	"log"
)

/**
Имеем массив с данными аккаунтов и кол-во персонажей на данном аккаунте.
Требуется на форме при выбора сервера
Условно пусть кол-во персонажей обновляется раз в N минут
*/

type List struct {
	ID      int    //ServerID
	Account string //Название аккаунта
	Count   int    //Кол-во персов
}

var AccountCount []List

/*
Чтение всех аккаунтов и персонажей
*/
func Get(db *pgx.Conn, ConnGS []*pgx.Conn) {
	accountList := getAccounts(db)
	getCharacters(ConnGS, accountList)
}

func getAccounts(db *pgx.Conn) []string {
	var accounts []string
	sql := `SELECT login FROM "accounts"`
	rows, err := db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var login string
		err = rows.Scan(&login)
		if err != nil {
			log.Println(err)
		}
		accounts = append(accounts, login)
	}
	return accounts
}

func getCharacters(conndb []*pgx.Conn, accountList []string) {
	sql := `SELECT login FROM "characters"`
	for id, db := range conndb {
		rows, err := db.Query(sql)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			var loginChar string
			err = rows.Scan(&loginChar)
			if err != nil {
				log.Println(err)
			}
			for _, login := range accountList {
				if login == loginChar {
					var isFindAccount = false
					for idacc, acc := range AccountCount {
						if acc.Account == login {
							if acc.ID == id {
								AccountCount[idacc].Count = AccountCount[idacc].Count + 1
								isFindAccount = true
							}
						}
					}
					if isFindAccount == false {
						AccountCount = append(AccountCount, List{
							ID:      id,
							Account: loginChar,
							Count:   1,
						})
					}
				}
			}
		}
	}

}

/**
Кол-во аккаунтов на персонаже
*/
func CountCharacterInAccount(sid int, login string) int {
	for _, account := range AccountCount {
		if account.ID == sid && account.Account == login {
			return account.Count
		}
	}
	return 0
}
