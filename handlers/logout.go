package handlers

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"itmo-ps-auth/database"
	"net/http"
	"time"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		p := jwt.Parser{
			ValidMethods:         nil,
			UseJSONNumber:        false,
			SkipClaimsValidation: false,
		}

		JWTCookie, _ := r.Cookie("JWT")
		token, _, _ := p.ParseUnverified(JWTCookie.Value, &jwt.StandardClaims{})
		login := token.Claims.(*jwt.StandardClaims).Subject

		db := database.Get("users")
		if db == nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		err := database.ExecCtx(ctx, db, "ALTER TABLE tokens DELETE WHERE login=?", login)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "JWT",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			HttpOnly: true,
			Secure: true,
		})

		http.SetCookie(w, &http.Cookie{
			Name:    "REFRESH",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			HttpOnly: true,
			Secure: true,
		})

		http.Redirect(w, r, "/signin", http.StatusSeeOther)

	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}