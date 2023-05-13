package main

import (
	"context"
	"os"
	"reflect"
	"sync"
	"time"

	v1networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ingresses = []Ingress{}
var ingressesmu sync.Mutex

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
		if condition := ingress.ObjectMeta.Name == dnsConfig.IngressName; condition {
			for _, rule := range ingress.Spec.TLS {
				domains = rule.Hosts
			}
		}
		ingressEvents <- Ingress{Name: ingress.Name, Domains: domains}
	}
}

func replaceIngressIfDifList(newIngress Ingress) (bool) {
	ingressesmu.Lock()
	for i, ingress := range ingresses {
		if reflect.DeepEqual(ingress, newIngress) {
			logger.Infow("No change in ingress data",
				"currentTime", time.Now(),
			)
			ingressesmu.Unlock()
			return false
		} else if ingress.Name == newIngress.Name {
			ingresses[i] = newIngress
			logger.Infow("Replaced ingress due to change in ingress data",
				"currentTime", time.Now(),
			)
			ingressesmu.Unlock()
			return true
		}
	}
	logger.Infow("New ingress added",
		"currentTime", time.Now(),
	)
	ingresses = append(ingresses, newIngress)
	ingressesmu.Unlock()
	return true
}
