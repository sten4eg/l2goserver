package accounts

import (
	"context"
	"l2goserver/db"
	"log"
)

// List Имеем массив с данными аккаунтов и кол-во персонажей на данном аккаунте.
// Требуется на форме при выборе сервера
// Условно пусть кол-во персонажей обновляется раз в N минут
type List struct {
	ID      int    //ServerID
	Account string //Название аккаунта
	Count   int    //Кол-во персов
}

var accountCount []List

// Get Чтение всех аккаунтов и персонажей
func Get() {
	accountList := getAccounts()
	getCharacters(accountList)
}

func getAccounts() []string {
	dbConn, err := db.GetConn()
	if err != nil {
		panic(err.Error())
	}
	defer dbConn.Release()

	var accounts []string
	sql := `SELECT login FROM "accounts"`
	rows, err := dbConn.Query(context.Background(), sql)
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

func getCharacters(accountList []string) {
	dbConn, err := db.GetConn()
	if err != nil {
		panic(err.Error())
	}
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
					for idacc, acc := range accountCount {
						if acc.Account == login {
							if acc.ID == id {
								accountCount[idacc].Count = accountCount[idacc].Count + 1
								isFindAccount = true
							}
						}
					}
					if isFindAccount == false {
						accountCount = append(accountCount, List{
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

// CountCharacterInAccount Кол-во аккаунтов на персонаже
func CountCharacterInAccount(sid int, login string) int {
	for _, account := range accountCount {
		if account.ID == sid && account.Account == login {
			return account.Count
		}
	}
	return 0
}
