package main

import (
	"context"
	"github.com/spf13/viper"
	"itmo-ps-auth/logger"
	"itmo-ps-auth/metrics"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	log = logger.New("main")
)

func main() {
	// get ENV variables
	viper.AutomaticEnv()

	// for interact on SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	srv := http.Server{
		Addr: viper.GetString("ADDR"),
		Handler: GetAPI(),
	}

	log.Infof("Starting listen on %v", viper.GetString("ADDR"))
	go listenHTTP(srv)
	go metrics.Collect()
	go metrics.DeleteOldMetrics()

	// wait for system signals
	<-signals

	log.Warnf("Receive SIGINT or SIGTERM. Starting shutdown")

	// try shutdown gracefully
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.WithError(err).Fatalf("Server shutdown failed")
	}

	log.Warnf("Shutdown complete")
}

func listenHTTP(srv http.Server) {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Fatalf("Error when listening")
	}
}
