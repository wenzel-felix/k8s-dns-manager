package main

import (
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var dnsConfig *DNSConfiguration
var kubernetesClient *kubernetes.Clientset

func init() {
	dnsConfig = InitDNSConfiguration()
	kubernetesClient = InitKubernetesClient()
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			logger.Infow("Failed to access required variable",
				"currentTime", time.Now(),
				"variableName", key,
			)
			os.Exit(1)
		}
		return defaultValue
	}
	return value
}

func InitDNSConfiguration() *DNSConfiguration {
	serverPort := getEnvOrDefault("TARGET_PORT", "80")
	protocol := getEnvOrDefault("TARGET_PROTOCOL", "http")
	path := getEnvOrDefault("TARGET_PATH", "/healthz")
	cloudflareDomain := getEnvOrDefault("CLOUDFLARE_DOMAIN", "")
	cloudflareToken := getEnvOrDefault("CLOUDFLARE_TOKEN", "")
	ingressName := getEnvOrDefault("INGRESS_NAME", "")

	return &DNSConfiguration{
		CloudflareDomain: cloudflareDomain,
		CloudflareToken:  cloudflareToken,
		TargetPort:       serverPort,
		TargetProtocol:   protocol,
		TargetPath:       path,
		IngressName:      ingressName,
	}
}

func InitKubernetesClient() *kubernetes.Clientset {
	config := &rest.Config{}
	initErr := error(nil)

	if condition := getEnvOrDefault("ENVIRONMENT", "PRD") == "DEV"; condition {
		// logger.Infow("client: using local kubeconfig",
		// 	"currentTime", time.Now(),
		// )
		config, initErr = clientcmd.BuildConfigFromFlags("", "kubeconfig")
		if initErr != nil {
			panic(initErr.Error())
		}
	} else {
		// logger.Infow("client: using in-cluster kubeconfig",
		// 	"currentTime", time.Now(),
		// )
		config, initErr = rest.InClusterConfig()
		if initErr != nil {
			panic(initErr.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
