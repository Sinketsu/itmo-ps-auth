package handlers

import (
	"github.com/spf13/viper"
	"github.com/yeka/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io"
	"itmo-ps-auth/database"
	"itmo-ps-auth/logger"
	"itmo-ps-auth/security"
	"net/http"
	"time"
)

func Download(w http.ResponseWriter, r *http.Request) {
	log := logger.New("Download")

	p := jwt.Parser{
		ValidMethods:         nil,
		UseJSONNumber:        false,
		SkipClaimsValidation: false,
	}

	JWTCookie, _ := r.Cookie("JWT")
	token, _, _ := p.ParseUnverified(JWTCookie.Value, &jwt.StandardClaims{})
	stdClaims := token.Claims.(*jwt.StandardClaims)
	role := stdClaims.Id

	if role != security.RoleAdmin {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

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

	var content bytes.Buffer
	content.WriteString(fmt.Sprintf("\"%v\";\"%v\";%v\n", "Series", "Time", "Value"))
	var zipped bytes.Buffer

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

		content.WriteString(fmt.Sprintf("\"%v\";\"%v\";%v\n", Type, Timestamp, Value))
	}

	zipw := zip.NewWriter(&zipped)
	defer zipw.Close()

	wr, err := zipw.Encrypt("data.csv", viper.GetString("ZIP_KEY"), zip.AES256Encryption)
	if err != nil {
		log.WithError(err).Errorf("Can't create encrypted writer")
	}

	_, err = io.Copy(wr, bytes.NewReader(content.Bytes()))
	if err != nil {
		log.WithError(err).Errorf("Can't write content")
	}
	zipw.Close()

	w.Write(zipped.Bytes())
}
