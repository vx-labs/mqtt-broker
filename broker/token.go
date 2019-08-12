package broker

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/vx-labs/mqtt-broker/sessions"
)

type Token struct {
	SessionID     string `json:"session_id"`
	PeerID        string `json:"peer_id"`
	SessionTenant string `json:"session_tenant"`
	jwt.StandardClaims
}

func EncodeSessionToken(signKey string, session sessions.Session) (string, error) {
	t := Token{
		SessionID:     session.ID,
		PeerID:        session.Peer,
		SessionTenant: session.Tenant,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(72 * time.Hour).Unix(),
			Issuer:    "mqtt-broker",
			Subject:   session.ID,
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
