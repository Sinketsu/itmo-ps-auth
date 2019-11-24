package handlers

import (
	"context"
	"database/sql"
	_ "github.com/kshvakov/clickhouse"
	"github.com/spf13/viper"
	"html/template"
	"itmo-ps-auth/database"
	"itmo-ps-auth/logger"
	"itmo-ps-auth/security"
	"net/http"
	"time"
)


func SignUp(w http.ResponseWriter, r *http.Request) {
	log := logger.New("SignUp")

	if r.Method == http.MethodGet {
		tmpl, err := template.New("signup.html").ParseFiles("frontend/signup.html")
		if err != nil {
			log.WithError(err).Errorf("Can't parse signup.html")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			log.WithError(err).Errorf("Can't execute signup.html")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.WithError(err).Errorf("Can't parse form")
			http.Error(w, "Bad form", http.StatusBadRequest)
			return
		}

		login := r.Form.Get("login")
		password := r.Form.Get("password")

		if len(login) == 0 {
			http.Error(w, "login is required", http.StatusBadRequest)
			return
		}

		if len(password) == 0 {
			http.Error(w, "password is required", http.StatusBadRequest)
			return
		}

		db := database.Get("users")
		if db == nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		row := db.QueryRowContext(ctx, "SELECT login FROM users WHERE login=?", login)
		var tempLogin string
		if err := row.Scan(&tempLogin); err == nil {
			http.Error(w, "User already registered", http.StatusConflict)
			return
		} else if err != sql.ErrNoRows {
			log.WithError(err).Errorf("Can't select users")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		hashed := security.HashPassword(password)
		token := security.NewRefreshToken()

		err = database.ExecCtx(ctx, db, "INSERT INTO users (created, login, password) VALUES (?, ?, ?)",
			time.Now(), login, hashed)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		err = database.ExecCtx(ctx, db, "INSERT INTO tokens (login, value, expired) VALUES (?, ?, ?)",
			login, token, time.Now().Add(viper.GetDuration("REFRESH_DURATION")))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		refreshCookie := &http.Cookie{
			Name:    "REFRESH",
			Value:   token,
			Path:    "/",
			Expires: time.Now().Add(viper.GetDuration("REFRESH_DURATION")),
			HttpOnly: true,
			Secure: true,
		}

		http.SetCookie(w, refreshCookie)
		err = security.UpdateJWT(w, login)
		if err != nil {
			log.WithError(err).Errorf("Can't update JWT")
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}
