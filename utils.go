package main

import (
	"fmt"
	"net/http"
	"time"
)

// returns true if the given host is available else it returns false
func VerifyAvailability(protocol string, host string, serverPort string, path string) bool {
	requestURL := fmt.Sprintf("%s://%s:%s%s", protocol, host, serverPort, path)

	logger.Infow("Verifying availability",
		"Target URL", requestURL,
	)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		logger.Errorw("Could not assable request",
			"Target URL", requestURL,
		)
		return false
	}

	client := http.Client{
		Timeout: 20 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.Infow("Node endpoint not availabe",
			"Target URL", requestURL,
			"Error", err,
		)
		return false
	}

	if condition := res.StatusCode == http.StatusOK; !condition {
		logger.Infow("Node endpoint responded with non-OK status code",
			"Target URL", requestURL,
			"HTTP-Code", res.StatusCode,
		)
		return false
	}

	logger.Infow("Node endpoint is available",
		"Target URL", requestURL,
	)
	return true
}
