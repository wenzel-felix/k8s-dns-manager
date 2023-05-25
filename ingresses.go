package main

import (
	"context"
	"time"

	v1networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// watchIngressData watches the Kubernetes API for changes to the ingress data
func watchIngressData(ingressEvents chan Ingress) {
	watch, err := kubernetesClient.NetworkingV1().Ingresses("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Fatalw("Error creating event watch for ingress data",
			"currentTime", time.Now(),
		)
	}
	defer watch.Stop()

	for event := range watch.ResultChan() {
		ingress := event.Object.(*v1networking.Ingress)
		domains := []string{}
		targets := []string{}

		if event.Type != "DELETED" {
			convertedIngress := convertIngress(ingress)
			logger.Infow("Received new ingress event",
				"ingress", ingress.Name,
				"ingressDomains", domains,
				"ingressTargets", targets,
			)
			ingressEvents <- convertedIngress
		} else {
			logger.Infow("Received deleted ingress event",
				"ingress", ingress.Name,
			)
			ingressEvents <- Ingress{Name: ingress.Name, Domains: domains, Targets: targets}
		}
	}
}

// createAIngressRoutine that gets all ingress resources and pushes them into a channel
func createIngressRoutine(currentIngresses chan Ingress) {
	for i := 0; ; i++ {
		time.Sleep(5 * time.Minute)
		ingresses, err := kubernetesClient.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			logger.Fatalw("Error retrieving ingress data",
				"error", err,
			)
		}
		for _, ingress := range ingresses.Items {
			convertedIngress := convertIngress(&ingress)
			logger.Infow("Received new ingress event",
				"ingress", convertedIngress.Name,
				"ingressDomains", convertedIngress.Domains,
				"ingressTargets", convertedIngress.Targets,
			)
			currentIngresses <- convertedIngress
		}
	}
}

// convert *v1networking.Ingress to Ingress
func convertIngress(ingress *v1networking.Ingress) Ingress {
	domains := []string{}
	targets := []string{}

	for _, rule := range ingress.Spec.Rules {
		domains = append(domains, rule.Host)
	}
	for _, target := range ingress.Status.LoadBalancer.Ingress {
		if target.IP != "" {
			targets = append(targets, target.IP)
		}
	}

	return Ingress{Name: ingress.Name, Domains: domains, Targets: targets}
}
