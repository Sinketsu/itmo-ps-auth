package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"itmo-ps-auth/logger"
)

var (
	log = logger.New("Database")
)

func Get(database string) *sql.DB {
	db, err := sql.Open("clickhouse",
		fmt.Sprintf("tcp://%v?username=kapmon&password=%v&database=%v",
			viper.GetString("CLICKHOUSE_ADDR"), viper.GetString("CLICKHOUSE_PASSWORD"),
			database))

	if err != nil {
		log.WithError(err).Errorf("Can't open database")
		return nil
	}

	return db
}

func ExecCtx(ctx context.Context, db *sql.DB, stmt string, args ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		logrus.WithError(err).Errorf("Can't begin transaction")
		return err
	}

	_, err = tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		logrus.WithError(err).Errorf("Can't execute stmt. Rolling back")
		errRollBack := tx.Rollback()
		if errRollBack != nil {
			logrus.WithError(errRollBack).Errorf("Can't rollback transaction")
			return errRollBack
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		logrus.WithError(err).Errorf("Can't commit transaction. Rolling back")
		errRollBack := tx.Rollback()
		if errRollBack != nil {
			logrus.WithError(errRollBack).Errorf("Can't rollback transaction")
			return errRollBack
		}
		return err
	}

	return nil
}
