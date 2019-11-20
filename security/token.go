package security

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/jmsleiman/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

func NewRefreshToken() string {
	r := make([]byte, 24)
	rand.Read(r)
	return base64.StdEncoding.EncodeToString(r)
}

func NewJWT(login string) (string, error) {
	authJWT := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(10 * time.Second).Unix(),
		Subject:   login,
	})

	return authJWT.SignedString([]byte(viper.GetString("JWT_SECRET")))
}

func UpdateJWT(w http.ResponseWriter, login string) error {
	accessToken, err := NewJWT(login)
	if err != nil {
		logrus.WithError(err).Errorf("Can't create new JWT token")
		return err
	}

	authCookie := &http.Cookie{
		Name:    "JWT",
		Value:   accessToken,
		Path:    "/",
		Expires: time.Now().Add(viper.GetDuration("REFRESH_DURATION")),
		HttpOnly: true,
		//Secure: true,
	}
	http.SetCookie(w, authCookie)

	return nil
}