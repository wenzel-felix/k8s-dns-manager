package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
)

func AdjustDNSEntries(ingresses []Ingress, nodeStatuses []NodeStatus, dnsConfig DNSConfiguration) ([]NodeStatus, bool) {
	api, err := cloudflare.NewWithAPIToken(dnsConfig.CloudflareToken)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(dnsConfig.CloudflareDomain)
	if err != nil {
		log.Fatal(err)
	}
	zoneIdentifier := cloudflare.ZoneIdentifier(zoneID)

	domains := []string{}
	for _, ingress := range ingresses {
		domains = append(domains, ingress.Domains...)
	}

	for i, host := range nodeStatuses {
		if (!host.EndpointAvailable || !host.KubeletAvailable) && host.DomainsConfigured {
			recs, _, err := api.ListDNSRecords(context.Background(),
				zoneIdentifier,
				cloudflare.ListDNSRecordsParams{Type: "A", Content: host.IP})
			if err != nil {
				log.Fatal(err)
			}
			for _, record := range recs {
				err := api.DeleteDNSRecord(context.Background(), zoneIdentifier, record.ID)
				if err != nil {
					log.Println(err)
				} else {
					fmt.Printf("A record deleted: %s => %s\n", record.Name, record.Content)
				}
			}
			continue
		}

		for _, domain := range domains {
			fmt.Printf("host: %s, endpoint available: %t, kubelet available: %t, status: %t\n", host.IP, host.EndpointAvailable, host.KubeletAvailable, host.EndpointAvailable || host.KubeletAvailable)
			recs, info, err := api.ListDNSRecords(context.Background(),
				zoneIdentifier,
				cloudflare.ListDNSRecordsParams{Type: "A", Content: host.IP, Name: domain})

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Existing records: %+v\n", info.Count)

			if condition := ((host.EndpointAvailable || host.KubeletAvailable) && info.Count == 0); condition {
				record, err := api.CreateDNSRecord(context.Background(), zoneIdentifier, cloudflare.CreateDNSRecordParams{
					Type:    "A",
					Name:    domain,
					Content: host.IP,
				})
				if err != nil {
					log.Println(err)
				} else {
					fmt.Printf("A record created: %s => %s\n", record.Name, record.Content)
					nodeStatuses[i].DomainsConfigured = true
				}
			} else if condition := (!host.EndpointAvailable && !host.KubeletAvailable && info.Count > 0); condition {
				for _, record := range recs {
					err := api.DeleteDNSRecord(context.Background(), zoneIdentifier, record.ID)
					if err != nil {
						log.Println(err)
					} else {
						fmt.Printf("A record deleted: %s => %s\n", record.Name, record.Content)
					}
				}
			} else if condition := ((host.EndpointAvailable || host.KubeletAvailable) && info.Count > 0); condition {
				fmt.Printf("A record already exists: %s => %s\n", domain, host.IP)
			}
		}
	}
	return nodeStatuses, false
}
