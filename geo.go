package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func getGeoInfo(ip string) (*GeoInfo, error) {
	// leafcloudhq/echoip docker container
	req, err := http.NewRequest("GET", "http://localhost:8080/json?ip="+ip, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info GeoInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return &info, err
}

func ipFromRequest(headers []string, r *http.Request) (net.IP, error) {
	remoteIP := ""
	for _, header := range headers {
		remoteIP = r.Header.Get(header)
		if http.CanonicalHeaderKey(header) == "X-Forwarded-For" {
			remoteIP = ipFromForwardedForHeader(remoteIP)
		}
		if remoteIP != "" {
			break
		}
	}

	if remoteIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return nil, err
		}
		remoteIP = host
	}

	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return nil, fmt.Errorf("could not parse IP: %s", remoteIP)
	}
	return ip, nil
}

func ipFromForwardedForHeader(v string) string {
	sep := strings.Index(v, ",")
	if sep == -1 {
		return v
	}
	return v[:sep]
}
