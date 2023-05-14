package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// a helper function to get environment variables or return a default value
func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			logger.Infow("Failed to access required variable",
				"variableName", key,
			)
			os.Exit(1)
		}
		return defaultValue
	}
	return value
}

// InitDNSConfiguration initializes the DNSConfiguration struct
func InitDNSConfiguration() {
	targetPort := getEnvOrDefault("TARGET_PORT", "80")
	protocol := getEnvOrDefault("TARGET_PROTOCOL", "http")
	path := getEnvOrDefault("TARGET_PATH", "/healthz")
	cloudflareDomain := getEnvOrDefault("CLOUDFLARE_DOMAIN", "")
	cloudflareToken := getEnvOrDefault("CLOUDFLARE_TOKEN", "")

	logger.Infow("Successfully retrieved environment variables", 
		"targetPort", targetPort,
		"protocol", protocol,
		"path", path,
		"cloudflareDomain", cloudflareDomain,
		"cloudflareToken", "********",
	)

	dnsConfig = &DNSConfiguration{
		CloudflareDomain: cloudflareDomain,
		CloudflareToken:  cloudflareToken,
		TargetPort:       targetPort,
		TargetProtocol:   protocol,
		TargetPath:       path,
	}

	logger.Infow("Successfully initialized DNSConfiguration struct")
}

// InitKubernetesClient initializes the Kubernetes client
func InitKubernetesClient() {
	config := &rest.Config{}
	initErr := error(nil)

	if condition := getEnvOrDefault("ENVIRONMENT", "PRD") == "DEV"; condition {
		logger.Infow("Using local kubeconfig")
		config, initErr = clientcmd.BuildConfigFromFlags("", "kubeconfig")
	} else {
		logger.Infow("Using in-cluster kubeconfig")
		config, initErr = rest.InClusterConfig()
	}

	if initErr != nil {
		logger.Panicw("Failed to retrieve Kubernetes client config", 
			"error", initErr,)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Panicw("Failed to initialize Kubernetes client", 
			"error", initErr,)
	}
	kubernetesClient = clientset
}
