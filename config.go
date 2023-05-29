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

// initDNSConfiguration initializes the DNSConfiguration struct
func initDNSConfiguration() {
	targetPort := getEnvOrDefault("TARGET_PORT", "80")
	protocol := getEnvOrDefault("TARGET_PROTOCOL", "http")
	path := getEnvOrDefault("TARGET_PATH", "/healthz")
	cloudflareToken := getEnvOrDefault("CLOUDFLARE_TOKEN", "")
	cloudflareTokenEnd  := cloudflareToken[len(cloudflareToken)-4:]

	logger.Infow("Successfully retrieved environment variables", 
		"targetPort", targetPort,
		"protocol", protocol,
		"path", path,
		"cloudflareToken", "********" + cloudflareTokenEnd,
	)

	dnsConfig = &DNSConfiguration{
		CloudflareToken:  cloudflareToken,
		TargetPort:       targetPort,
		TargetProtocol:   protocol,
		TargetPath:       path,
	}

	logger.Infow("Successfully initialized DNSConfiguration struct")
}

// initKubernetesClient initializes the Kubernetes client
func initKubernetesClient() {
	config := &rest.Config{}
	initErr := error(nil)

	logger.Infow("Using in-cluster kubeconfig")
	config, initErr = rest.InClusterConfig()

	if initErr != nil {
		logger.Infow("Failed to retrieve in-cluster Kubernetes client config. Falling back to local kubeconfig",
			"error", initErr,
		)
		config, initErr = clientcmd.BuildConfigFromFlags("", "kubeconfig")
		if initErr != nil {
			logger.Panicw("Failed to retrieve local Kubernetes client config",
				"error", initErr,
			)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Panicw("Failed to initialize Kubernetes client", 
			"error", initErr,)
	}
	kubernetesClient = clientset
}
