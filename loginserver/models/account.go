package models



type Account struct {
//TODO WTF	Id          bson.ObjectId `bson:"_id,omitempty"`
	Username    string        `bson:"username"`
	Password    string        `bson:"password"`
	AccessLevel int8          `bson:"access_level"`
}