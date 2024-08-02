package cookies

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func Write(w http.ResponseWriter, r *http.Request, secretKeyData []byte, userID int64, cookie http.Cookie) error {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"Sub":       userID,
			"Issued":    time.Now(),
			"NotBefore": time.Now(),
			"Expires":   time.Now().Add(24 * time.Hour),
			"Issuer":    "StratCheck",
			"Audience":  []string{"BacktestingApi"},
		})

	secretKey, err := jwt.ParseRSAPrivateKeyFromPEM(secretKeyData)
	if err != nil {
		return err
	}

	tokenString, err := jwtToken.SignedString(secretKey)
	if err != nil {
		return err
	}
	if len(tokenString) > 4096 {
		return ErrValueTooLong
	}

	cookie.Value = tokenString
	http.SetCookie(w, &cookie)

	return nil
}

func Read(r *http.Request, name string, secretKey []byte) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
