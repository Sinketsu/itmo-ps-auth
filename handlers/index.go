package handlers

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"html/template"
	"itmo-ps-auth/database"
	"itmo-ps-auth/logger"
	"itmo-ps-auth/security"
	"net/http"
	"time"
)

type DataSeries struct {
	Timestamp time.Time
	Value float64
}

type Result struct {
	Login string
	Role string
	Data [][]DataSeries
}

func Index(w http.ResponseWriter, r *http.Request) {
	log := logger.New("Index")

	db := database.Get("stats")
	if db == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT * FROM stats WHERE timestamp >= (now() - toIntervalMinute(10)) ORDER by timestamp")
	if err != nil {
		log.WithError(err).Errorf("Can't select metrics")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer rows.Close()

	var cpu, memory, la5 []DataSeries

	var Timestamp time.Time
	var Type string
	var Value float64
	for rows.Next() {
		err := rows.Scan(&Timestamp, &Type, &Value)
		if err != nil {
			log.WithError(err).Errorf("Can't scan row")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		switch Type {
		case "cpu":
			cpu = append(cpu, DataSeries{Timestamp, Value})
		case "memory":
			memory = append(memory, DataSeries{Timestamp, Value / 1024 / 1024})
		case "la5":
			la5 = append(la5, DataSeries{Timestamp, Value})
		}
	}

	tmpl, err := template.New("index.html").ParseFiles("frontend/index.html")
	if err != nil {
		log.WithError(err).Errorf("Can't parse index.html")
		http.Error(w, "Template error", http.StatusServiceUnavailable)
		return
	}

	p := jwt.Parser{
		ValidMethods:         nil,
		UseJSONNumber:        false,
		SkipClaimsValidation: false,
	}

	JWTCookie, _ := r.Cookie("JWT")
	token, _, _ := p.ParseUnverified(JWTCookie.Value, &jwt.StandardClaims{})
	stdClaims := token.Claims.(*jwt.StandardClaims)
	login := stdClaims.Subject
	role := stdClaims.Id

	result := &Result{
		Login: login,
		Role: role,
		Data:  [][]DataSeries{cpu, memory, la5},
	}
	if role == security.RoleServant {
		result.Data = [][]DataSeries{cpu}
	}

	err = tmpl.Execute(w, result)
	if err != nil {
		log.WithError(err).Errorf("Can't execute index.html")
		http.Error(w, "Template error", http.StatusServiceUnavailable)
		return
	}
}
