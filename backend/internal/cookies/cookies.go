package cookies

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	ErrValueTooLong      = errors.New("cookie value too long")
	ErrInvalidValue      = errors.New("invalid cookie value")
	ErrInvalidPrivateKey = errors.New("failed to decode PEM block containing private key")
)

func Write(w http.ResponseWriter, r *http.Request, secretKeyData []byte, userID int64, cookie *http.Cookie) error {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			"sub": userID,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iss": "StratCheck",
			"aud": []string{"BacktraderAPI"},
		})

	block, _ := pem.Decode(secretKeyData)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return ErrInvalidPrivateKey
	}

	secretKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	tokenString, err := jwtToken.SignedString(secretKey)
	if err != nil {
		return err
	}

	if len(tokenString) > 512 {
		return ErrValueTooLong
	}

	cookie.Value = tokenString
	http.SetCookie(w, cookie)

	return nil
}

func Read(r *http.Request, name string, secretKey []byte) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
