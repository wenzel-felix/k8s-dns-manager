package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// returns true if the given string is a valid IPv4 address else it returns false
func IsValidIPv4Address(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	if parsedIP.To4() == nil {
		return false
	}
	return true
}

// returns true if the given host is available else it returns false
func VerifyAvailability(protocol string, host string, serverPort string, path string) bool {
	requestURL := fmt.Sprintf("%s://%s:%s%s", protocol, host, serverPort, path)

	logger.Infow("Verifying availability",
				"Target URL", requestURL,
			)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return false
	}
	fmt.Printf("client: created request!\n")
	
	client := http.Client{
		Timeout: 20 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return false
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	if condition := res.StatusCode == http.StatusOK; !condition {
		fmt.Printf("client: status code not OK: %d\n", res.StatusCode)
		return false
	}
	return true
}
