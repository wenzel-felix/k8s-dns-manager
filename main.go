package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

type Host struct {
	IP        string
	Available bool
}

func main() {
	serverPort := getEnvOrDefault("TARGET_PORT", "8000")
	// https://iximiuz.com/en/posts/kubernetes-api-call-simple-http-client/
	hostsString := getEnvOrDefault("TARGET_IPS", "localhost,127.0.0.1,10.10.1.5")
	hostsArray := strings.Split(hostsString, ",")
	protocol := getEnvOrDefault("TARGET_PROTOCOL", "http")
	path := getEnvOrDefault("TARGET_PATH", "/healthz")
	cloudflareDomain := getEnvOrDefault("CLOUDFLARE_DOMAIN", "")
	cloudflareToken := getEnvOrDefault("CLOUDFLARE_TOKEN", "")

	var hostStatuses []Host

	for _, host := range hostsArray {
		if condition := isValidIPv4Address(host); condition {
			status := verifyAvailability(protocol, host, serverPort, path)
			hostStatuses = append(hostStatuses, Host{IP: host, Available: status})
		} else {
			fmt.Printf("Skipping host: %s as it is not a valid IPv4 address.\n", host)
		}
	}

	api, err := cloudflare.NewWithAPIToken(cloudflareToken)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(cloudflareDomain)
	if err != nil {
		log.Fatal(err)
	}

	for _, host := range hostStatuses {
		fmt.Printf("host: %s, available: %t\n", host.IP, host.Available)
		recs, info, err := api.ListDNSRecords(context.Background(),
			cloudflare.ZoneIdentifier(zoneID),
			cloudflare.ListDNSRecordsParams{Type: "A", Content: host.IP, Name: fmt.Sprintf("*.%s", cloudflareDomain)})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Existing records: %+v\n", info.Count)

		if condition := (host.Available && info.Count == 0); condition {
			record, err := api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
				Type:    "A",
				Name:    "*",
				Content: host.IP,
			})
			if err != nil {
				log.Println(err)
			} else {
				fmt.Printf("A record created: %s => %s\n", record.Name, record.Content)
			}
		} else if condition := (!host.Available && info.Count > 0); condition {
			for _, record := range recs {
				err := api.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), record.ID)
				if err != nil {
					log.Println(err)
				} else {
					fmt.Printf("A record deleted: %s => %s\n", record.Name, record.Content)
				}
			}
		} else if condition := (host.Available && info.Count > 0); condition {
			fmt.Printf("A record already exists: %s.%s => %s\n", "*", cloudflareDomain, host.IP)
		}
	}
}

func verifyAvailability(protocol string, host string, serverPort string, path string) bool {
	requestURL := fmt.Sprintf("%s://%s:%s%s", protocol, host, serverPort, path)

	fmt.Printf("client: making request to %s\n", requestURL)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return false
	}

	res, err := http.DefaultClient.Do(req)
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

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			fmt.Printf("No value for required variable %s provided\n", key)
			os.Exit(1)
		}
		return defaultValue
	}
	return value
}

func isValidIPv4Address(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	if parsedIP.To4() == nil {
		return false
	}
	return true
}
