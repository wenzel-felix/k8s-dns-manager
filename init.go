package main

import (
	"github.com/cloudflare/cloudflare-go"
	"k8s.io/client-go/kubernetes"
)

var cloudflareClient *cloudflare.API
var zoneIdentifier *cloudflare.ResourceContainer
var dnsConfig *DNSConfiguration
var kubernetesClient *kubernetes.Clientset

func init() {
	initLogger()
	InitDNSConfiguration()
	initCloudflareClient()
	InitKubernetesClient()
}