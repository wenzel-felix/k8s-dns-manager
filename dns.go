package main

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"golang.org/x/exp/slices"
	"log"
)

// initCloudflareClient initializes the Cloudflare client
func initCloudflareClient() {
	var err error
	cloudflareClient, err = cloudflare.NewWithAPIToken(dnsConfig.CloudflareToken)
	if err != nil {
		logger.Fatalw("Error initializing Cloudflare client",
			"error", err,
		)
	}
	zoneID, err := cloudflareClient.ZoneIDByName(dnsConfig.CloudflareDomain)
	if err != nil {
		logger.Fatalw("Error retrieving zone ID",
			"error", err,
		)
	}
	zoneIdentifier = cloudflare.ZoneIdentifier(zoneID)
	logger.Infow("Initialized Cloudflare client",
		"domain", dnsConfig.CloudflareDomain,
	)
}

// generateDNSRecordMap returns a map of IP addresses (targets) to DNS records
func getDNSRecordMaps(uniqueClusterComment string) (map[string]map[string]cloudflare.DNSRecord, map[string]map[string]cloudflare.DNSRecord) {
	// Map of hosts => DNS records name => DNS records
	hostNameRecordMap := make(map[string]map[string]cloudflare.DNSRecord)
	nameHostRecordMap := make(map[string]map[string]cloudflare.DNSRecord)

	recs, info, err := cloudflareClient.ListDNSRecords(context.Background(),
		zoneIdentifier,
		cloudflare.ListDNSRecordsParams{Type: "A",
			Comment: uniqueClusterComment})
	if err != nil {
		log.Fatal(err)
	}
	logger.Infow("Retrieved DNS records",
		"count", info.Count,
	)
	for _, record := range recs {
		if _, ok := hostNameRecordMap[record.Content]; !ok {
			hostNameRecordMap[record.Content] = make(map[string]cloudflare.DNSRecord)
		}
		hostNameRecordMap[record.Content][record.Name] = record

		if _, ok := nameHostRecordMap[record.Name]; !ok {
			nameHostRecordMap[record.Name] = make(map[string]cloudflare.DNSRecord)
		}
		nameHostRecordMap[record.Name][record.Content] = record
	}
	return hostNameRecordMap, nameHostRecordMap
}

// adjustDNSEntries adjusts the DNS entries in Cloudflare
func adjustDNSEntries(ingress Ingress, clusterUID string) bool {
	uniqueClusterComment := (clusterUID + ", " + ingress.Name)
	if len(uniqueClusterComment) > 50 {
		uniqueClusterComment = uniqueClusterComment[:50]
	}
	dnshostNameRecordMap, nameHostRecordMap := getDNSRecordMaps(uniqueClusterComment)

	for _, host := range ingress.Targets {
		hostMap, hostExists := dnshostNameRecordMap[host]
		if condition := (!hostExists); condition {
			createOrUpdateDNSRecords(ingress.Domains, []string{host}, uniqueClusterComment)
		} else {
			for _, domain := range ingress.Domains {
				_, hostDomainExists := hostMap[domain]

				if condition := (!hostDomainExists); condition {
					createOrUpdateDNSRecords([]string{domain}, []string{host}, uniqueClusterComment)
				} else if condition := (hostDomainExists); condition {
					logger.Infow("A record already exists",
						"domain", domain,
						"host", host,
					)
				}
			}
		}
	}

	for host, hostDomains := range dnshostNameRecordMap {
		if !slices.Contains(ingress.Targets, host) {
			for domain, rec := range hostDomains {
				logger.Infow("Deleting A record",
					"domain", domain,
					"host", host,
				)
				err := cloudflareClient.DeleteDNSRecord(context.Background(), zoneIdentifier, rec.ID)
				if err != nil {
					logger.Errorw("Error deleting DNS record",
						"error", err,
					)
				}
			}

		}
	}

	for domain, domainHosts := range nameHostRecordMap {
		if !slices.Contains(ingress.Domains, domain) {
			for host, rec := range domainHosts {
				logger.Infow("Deleting A record",
					"domain", domain,
					"host", host,
				)
				err := cloudflareClient.DeleteDNSRecord(context.Background(), zoneIdentifier, rec.ID)
				if err != nil {
					logger.Errorw("Error deleting DNS record",
						"error", err,
					)
				}
			}
		}
	}

	return false
}

func createOrUpdateDNSRecords(names []string, contents []string, comment string) {
	for _, name := range names {
		for _, content := range contents {
			_, err := cloudflareClient.CreateDNSRecord(context.Background(), zoneIdentifier, cloudflare.CreateDNSRecordParams{
				Type:    "A",
				Name:    name,
				Content: content,
				Comment: comment,
			})
			if err != nil {
				recs, _, err := cloudflareClient.ListDNSRecords(context.Background(), zoneIdentifier, cloudflare.ListDNSRecordsParams{
					Type:    "A",
					Name:    name,
					Content: content,
				})
				if err != nil {
					logger.Errorw("Error syncing DNS record",
						"error", err,
					)
				} else {
					_, err := cloudflareClient.UpdateDNSRecord(context.Background(),
						zoneIdentifier,
						cloudflare.UpdateDNSRecordParams{Type: "A", Name: name, Content: content, Comment: comment, ID: recs[0].ID})
					if err != nil {
						logger.Errorw("Error updating DNS record",
							"error", err,
						)
					}
					logger.Infow("A record updated",
						"domain", name,
						"host", content,
					)
				}
			} else {
				logger.Infow("A record created",
					"domain", name,
					"host", content,
				)
			}
		}
	}
}
