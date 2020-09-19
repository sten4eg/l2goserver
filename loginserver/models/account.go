package models

import (
	"database/sql"
	"github.com/jackc/pgx/pgtype"
)

type Account struct {
	Login       string
	Password    string
	CreatedAt   pgtype.Timestamp
	LastActive  pgtype.Timestamp
	AccessLevel int8
	LastIp      sql.NullString
	LastServer  int8
}
