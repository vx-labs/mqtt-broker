package auth

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	SessionID     string `json:"session_id"`
	SessionEntity string `json:"session_entity"`
	SessionTenant string `json:"session_tenant"`
	jwt.StandardClaims
}

func SigningKey() string {
	return os.Getenv("JWT_SIGN_KEY")
}

func EncodeSessionToken(signKey, tenant, entity, id string) (string, error) {
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
func DecodeSessionToken(signKey, signedToken string) (Token, error) {
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
