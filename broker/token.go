package broker

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	SessionID     string `json:"session_id"`
	SessionTenant string `json:"session_tenant"`
	jwt.StandardClaims
}

func EncodeSessionToken(signKey string, tenant string, id string) (string, error) {
	t := Token{
		SessionID:     id,
		SessionTenant: tenant,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(72 * time.Hour).Unix(),
			Issuer:    "mqtt-broker",
			Subject:   id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, t)
	return token.SignedString([]byte(signKey))
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
