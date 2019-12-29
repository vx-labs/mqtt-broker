package broker

import (
	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	SessionID     string `json:"session_id"`
	SessionTenant string `json:"session_tenant"`
	jwt.StandardClaims
}

func DecodeSessionToken(signKey string, signedToken string) (Token, error) {
	token, err := jwt.ParseWithClaims(signedToken, &Token{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})
	if err != nil {
		return Token{}, err
	}
	if claims, ok := token.Claims.(*Token); ok && token.Valid {
		return *claims, nil
	}
	return Token{}, err
}
