package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetEnvOrDefault(key string, defaultValue string) string {
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

func InitDNSConfiguration() DNSConfiguration {
	serverPort := GetEnvOrDefault("TARGET_PORT", "80")
	protocol := GetEnvOrDefault("TARGET_PROTOCOL", "http")
	path := GetEnvOrDefault("TARGET_PATH", "/healthz")
	cloudflareDomain := GetEnvOrDefault("CLOUDFLARE_DOMAIN", "")
	cloudflareToken := GetEnvOrDefault("CLOUDFLARE_TOKEN", "")
	ingressName := GetEnvOrDefault("INGRESS_NAME", "")

	return DNSConfiguration{
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

	if condition := GetEnvOrDefault("ENVIRONMENT", "PRD") == "DEV"; condition {
		fmt.Printf("client: using local kubeconfig\n")
		config, initErr = clientcmd.BuildConfigFromFlags("", "kubeconfig")
		if initErr != nil {
			panic(initErr.Error())
		}
	} else {
		fmt.Printf("client: using in-cluster kubeconfig\n")
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