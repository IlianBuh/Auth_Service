package models

type User struct {
	UUID     uint64
	Login    string
	Email    string
	PassHash []byte
}
