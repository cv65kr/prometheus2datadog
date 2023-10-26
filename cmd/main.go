package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cv65kr/prometheus2datadog/internal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	prometheusEndpoint := flag.String("metrics-endpoint", "http://0.0.0.0:2112", "Endpoint for scrapping e.g. 'http://localhost:2021'. Should contains prometheus format metrics")
	statsdAddr := flag.String("statsd-address", "0.0.0.0:8125", "Address for statsd e.g. 'localhost:8125'")
	scrapingInterval := flag.Int("scraping-interval", 60, "Interval for metrics scraping in seconds")
	shutdownTimeout := flag.Int("shutdown-timeout", 60, "Gracefull shutdown timeout in seconds")
	logLevel := flag.String("log-level", "debug", "log level debug, info, warn, error, fatal or panic")
	excludedMetrics := flag.String("exclude", "", "Prefix for excluded metrics e.g. test_,xxx_,www_")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] argument ...\n", os.Args[0])

		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args[1:]) == 1 {
		cmd := os.Args[1]
		if cmd == "help" {
			flag.Usage()
			os.Exit(1)
		}
	}

	logger, err := initZap(*logLevel)
	if err != nil {
		fmt.Println("Error during logger initialisation")
		os.Exit(1)
	}
	defer logger.Sync()

	if *prometheusEndpoint == "" {
		logger.Error("'prometheus-endpoint' field is required")
		os.Exit(1)
	}

	if *statsdAddr == "" {
		logger.Error("'statsd-address' field is required")
		os.Exit(1)
	}

	config := internal.Config{
		Interval:           time.Duration(*scrapingInterval) * time.Second,
		PrometheusEndpoint: *prometheusEndpoint,
		ExcludedMetrics:    strings.Split(*excludedMetrics, ","),
	}

	scraper := internal.NewMetricsScraper()
	datadogClient, err := internal.NewDatadogClient(*statsdAddr)
	if err != nil {
		logger.Error("creation datadog client", zap.Error(err))
		os.Exit(1)
	}
	converter := internal.NewConverter(datadogClient, logger, config)

	stopCh := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		close(stopCh)
		<-c
		os.Exit(1)
	}()

	worker := internal.NewWorker(scraper, datadogClient, converter, logger, config)
	worker.Run(stopCh)

	<-stopCh
	logger.Info("Received signal to kill")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(*shutdownTimeout)*time.Second)
	defer cancel()
	worker.Close()
}

func initZap(logLevel string) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	}

	zapEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zapEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}
