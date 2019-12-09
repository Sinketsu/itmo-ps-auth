package metrics

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/sirupsen/logrus"
	"itmo-ps-auth/database"
	"itmo-ps-auth/logger"
	"runtime"
	"time"
)

func Collect() {
	log := logger.New("Metrics")
	ticker := time.NewTicker(10 * time.Second)

	for {
		<- ticker.C

		CollectLA(log)
		CollectCPU(log)
		CollectMemory(log)
	}
}

func CollectLA(log *logrus.Entry) {
	la, err := load.Avg()
	if err != nil {
		log.WithError(err).Errorf("Can't collect LA")
		return
	}

	db := database.Get("stats")
	err = database.ExecCtx(context.Background(), db,
		"INSERT INTO stats (timestamp, type, value) VALUES (?, ?, ?)",
		time.Now(), "la5", la.Load5)

	if err != nil {
		log.WithError(err).Errorf("Can't insert new LA")
		return
	}
}

func CollectCPU(log *logrus.Entry) {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		log.WithError(err).Errorf("Can't collect CPU")
		return
	}

	db := database.Get("stats")
	err = database.ExecCtx(context.Background(), db,
		"INSERT INTO stats (timestamp, type, value) VALUES (?, ?, ?)",
		time.Now(), "cpu", percent[0])

	if err != nil {
		log.WithError(err).Errorf("Can't insert new CPU")
		return
	}
}

func CollectMemory(log *logrus.Entry) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	db := database.Get("stats")
	err := database.ExecCtx(context.Background(), db,
		"INSERT INTO stats (timestamp, type, value) VALUES (?, ?, ?)",
		time.Now(), "memory", float64(m.Alloc))

	if err != nil {
		log.WithError(err).Errorf("Can't insert new Memory")
		return
	}
}

func DeleteOldMetrics() {
	log := logger.New("Metrics")
	ticker := time.NewTicker(10 * time.Minute)

	for {
		<- ticker.C

		db := database.Get("stats")
		err := database.ExecCtx(context.Background(), db,
			"ALTER TABLE stats DELETE WHERE timestamp <= (now() - toIntervalMinute(60))")

		if err != nil {
			log.WithError(err).Errorf("Can't delete old metrics")
			return
		}
	}
}
