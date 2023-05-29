package main

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"golang.org/x/exp/slices"
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
	zones, err = cloudflareClient.ListZones(context.Background())
	if err != nil {
		logger.Fatalw("Error retrieving zone ID",
			"error", err,
		)
	}
	var zoneNames = []string{}
	for _, zone := range zones {
		zoneNames = append(zoneNames, zone.Name)
	}

	logger.Infow("Initialized Cloudflare client",
		"zones", zoneNames,
	)
}

// generateDNSRecordMap returns a map of IP addresses (targets) to DNS records
func getDNSRecordMaps(uniqueClusterComment string, zoneIdentifier *cloudflare.ResourceContainer, zoneName string) (map[string]map[string]cloudflare.DNSRecord, map[string]map[string]cloudflare.DNSRecord) {
	// Map of hosts => DNS records name => DNS records
	hostNameRecordMap := make(map[string]map[string]cloudflare.DNSRecord)
	nameHostRecordMap := make(map[string]map[string]cloudflare.DNSRecord)

	recs, info, err := cloudflareClient.ListDNSRecords(context.Background(),
		zoneIdentifier,
		cloudflare.ListDNSRecordsParams{Type: "A",
			Comment: uniqueClusterComment})
	if err != nil {
		logger.Fatalw("Error retrieving DNS records",
			"zone", zoneName,
			"error", err,
		)
	}
	logger.Infow("Retrieved DNS records",
		"zone", zoneName,
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
func adjustDNSZonesOrchestrator(ingress Ingress, clusterUID string) {
	uniqueClusterComment := (clusterUID + ", " + ingress.Name)
	if len(uniqueClusterComment) > 50 {
		uniqueClusterComment = uniqueClusterComment[:50]
	}

	for _, zone := range zones {
		go adjustDNSZone(zone, ingress, uniqueClusterComment)
	}
}

// check if subdomain is part of domain
func isSubdomain(domain string, subdomain string) bool {
	return domain == subdomain || strings.HasSuffix(subdomain, "."+domain)
}

func adjustDNSZone(zone cloudflare.Zone, ingress Ingress, uniqueClusterComment string) {
	zoneIdentifier := cloudflare.ZoneIdentifier(zone.ID)
	dnshostNameRecordMap, nameHostRecordMap := getDNSRecordMaps(uniqueClusterComment, zoneIdentifier, zone.Name)

	for _, host := range ingress.Targets {
		hostMap, hostExists := dnshostNameRecordMap[host]
		if condition := (!hostExists); condition {
			createOrUpdateDNSRecords(ingress.Domains, []string{host}, uniqueClusterComment, zoneIdentifier, zone.Name)
		} else {
			for _, domain := range ingress.Domains {
				_, hostDomainExists := hostMap[domain]

				if condition := (!hostDomainExists); condition {
					createOrUpdateDNSRecords([]string{domain}, []string{host}, uniqueClusterComment, zoneIdentifier, zone.Name)
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
				deleteDNSRecord(domain, host, zoneIdentifier, rec.ID)
			}
		}
	}

	for domain, domainHosts := range nameHostRecordMap {
		if !slices.Contains(ingress.Domains, domain) {
			for host, rec := range domainHosts {
				deleteDNSRecord(domain, host, zoneIdentifier, rec.ID)
			}
		}
	}
}

func deleteDNSRecord(domain string, host string, zoneIdentifier *cloudflare.ResourceContainer, recordId string) {
	logger.Infow("Deleting A record",
		"domain", domain,
		"host", host,
	)
	err := cloudflareClient.DeleteDNSRecord(context.Background(), zoneIdentifier, recordId)
	if err != nil {
		logger.Errorw("Error deleting DNS record",
			"error", err,
		)
	}
}

func createOrUpdateDNSRecords(names []string, contents []string, comment string, zoneIdentifier *cloudflare.ResourceContainer, zoneName string) {
	for _, name := range names {
		for _, content := range contents {
			if isSubdomain(zoneName, name) {
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
}
