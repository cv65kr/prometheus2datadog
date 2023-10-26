package internal

import (
	"time"

	"go.uber.org/zap"
)

type worker struct {
	scraper        scraper
	datadogClient  *datadogClient
	converter      converter
	logger         *zap.Logger
	config         Config
	confirmationCh chan bool
}

func NewWorker(scraper scraper, datadogClient *datadogClient, converter converter, logger *zap.Logger, config Config) worker {
	return worker{
		scraper:        scraper,
		datadogClient:  datadogClient,
		converter:      converter,
		logger:         logger,
		config:         config,
		confirmationCh: make(chan bool, 1),
	}
}
func (w worker) Run(stopCh <-chan struct{}) {
	go func() {
		for {
			w.logger.Info("Scrapping")
			metrics, err := w.scraper.Scrap(w.config.PrometheusEndpoint)
			if err != nil {
				w.logger.Error("Scraping", zap.Error(err))
				time.Sleep(w.config.Interval)
				continue
			}

			w.converter.Convert(metrics)
			w.datadogClient.Flush()
			w.logger.Debug("Metrics flush")

			select {
			case <-stopCh:
				w.confirmationCh <- true
				return
			default:
				w.logger.Info("Waiting for next round of scraping")
				time.Sleep(w.config.Interval)
			}
		}
	}()
}

func (w worker) Close() {
	w.logger.Info("Waiting for job finish")
	<-w.confirmationCh
	w.logger.Info("Close connection with Datadog")
	w.datadogClient.CloseConnection()
}
