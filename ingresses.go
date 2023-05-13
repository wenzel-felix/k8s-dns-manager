package main

import (
	"context"
	"fmt"
	"reflect"

	v1networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func watchIngressData(clientset *kubernetes.Clientset, dnsConfig DNSConfiguration, ingressEvents chan Ingress) {
	watch, err := clientset.NetworkingV1().Ingresses("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error creating watch: %s\n", err)
		return
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

func replaceIngressIfDifList(newIngress Ingress, ingresses []Ingress) ([]Ingress, bool) {
	for i, ingress := range ingresses {
		if reflect.DeepEqual(ingress, newIngress) {
			return ingresses, false
		} else if ingress.Name == newIngress.Name {
			ingresses[i] = newIngress
			return ingresses, true
		}
	}
	return append(ingresses, newIngress), true
}