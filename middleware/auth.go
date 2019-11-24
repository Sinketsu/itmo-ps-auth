package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"itmo-ps-auth/database"
	"itmo-ps-auth/logger"
	"itmo-ps-auth/security"
	"net/http"
	"time"
)

var log = logger.New("Auth")

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth, err := r.Cookie("JWT")
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		refresh, err := r.Cookie("REFRESH")
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		token, err := jwt.Parse(auth.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(viper.GetString("JWT_SECRET")), nil
		})

		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
					login, alive := CheckRefreshToken(refresh.Value)

					if alive {
						err = security.UpdateJWT(w, login)
						if err != nil {
							log.WithError(err).Errorf("Can't update JWT")
							http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
							return
						}

						next.ServeHTTP(w, r)
					} else {
						http.Redirect(w, r, "/signin", http.StatusSeeOther)
						return
					}
				} else {
					http.Redirect(w, r, "/signin", http.StatusSeeOther)
					return
				}
			} else {
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
				return
			}
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			next.ServeHTTP(w, r)
		}
	})
}

func CheckRefreshToken(token string) (string, bool) {
	db := database.Get("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	row := db.QueryRowContext(ctx, "SELECT login, expired FROM tokens WHERE value=?", token)

	var login string
	var expired time.Time
	if err := row.Scan(&login, &expired); err != nil {
		if err == sql.ErrNoRows {
			return "", false
		} else {
			log.WithError(err).Errorf("Can't select tokens")
			return "", false
		}
	}

	return login, time.Now().Before(expired)
}
