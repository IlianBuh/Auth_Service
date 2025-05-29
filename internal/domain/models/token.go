package models

const (
	RefreshToken = iota
	AccessToken
)

type Token struct {
	Type int
	Val  string
}

type TokensPair struct {
	RefreshToken Token
	AccessToken  Token
}
