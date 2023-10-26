package internal

import (
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type scraper struct{}

func NewMetricsScraper() scraper {
	return scraper{}
}

func (s scraper) Scrap(endpoint string) (map[string]*dto.MetricFamily, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, err
	}

	return mf, nil
}
