package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tracker"
)

func getMetric(what tracker.QueryType) ([]tracker.Metric, error) {
	data := tracker.MetricData{
		What:   what,
		SiteID: siteID,
		Start:  uint32(start),
		End:    uint32(end),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	host := tracker.GetConfig().GoTrackerHost
	req, err := http.NewRequest("POST", host+"/stats", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", tracker.GetConfig().APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var metrics []tracker.Metric
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println("error from API: ", string(b))
		return nil, err
	}

	return metrics, nil
}
