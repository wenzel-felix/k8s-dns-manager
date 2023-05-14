package main

import (
	"context"
	"os"
	"time"

	v1networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// watchIngressData watches the Kubernetes API for changes to the ingress data
func watchIngressData(ingressEvents chan Ingress) {
	watch, err := kubernetesClient.NetworkingV1().Ingresses("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Infow("Error creating event watch for ingress data",
			"currentTime", time.Now(),
		)
		os.Exit(1)
	}
	defer watch.Stop()

	for event := range watch.ResultChan() {
		ingress := event.Object.(*v1networking.Ingress)
		domains := []string{}
		targets := []string{}

		for _, rule := range ingress.Spec.TLS {
			domains = rule.Hosts
		}

		for _, target := range ingress.Status.LoadBalancer.Ingress {
			if target.IP != "" {
				targets = append(targets, target.IP)
			}
		}

		ingressEvents <- Ingress{Name: ingress.Name, Domains: domains, Targets: targets}
	}
}
