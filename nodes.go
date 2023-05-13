package main

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func watchNodeData(clientset *kubernetes.Clientset, dnsConfig DNSConfiguration, nodeEvents chan NodeStatus) {
	watch, err := clientset.CoreV1().Nodes().Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error creating watch: %s\n", err)
		nodeEvents <- NodeStatus{IP: "", EndpointAvailable: false}
	}
	defer watch.Stop()

	for event := range watch.ResultChan() {
		node := event.Object.(*v1.Node)
		if filteredNode, condition := filterNode(*node, dnsConfig); condition {
			fmt.Printf("Node: %s is ready\n", node.ObjectMeta.Name)
			nodeEvents <- filteredNode
		} else {
			fmt.Printf("Node: %s is not ready\n", node.ObjectMeta.Name)
			nodeEvents <- filteredNode
		}
	}
}

func filterNode(node v1.Node, dnsConfig DNSConfiguration) (NodeStatus, bool) {
	filteredNode := NodeStatus{IP: "", EndpointAvailable: false, KubeletAvailable: false, DomainsConfigured: false}
	for _, condition := range node.Status.Conditions {
		if condition.Reason == "KubeletReady" {
			if condition.Type == "Ready" && condition.Status == "False" {
				return filteredNode, false
			} else {
				filteredNode.KubeletAvailable = true
			}
		}
	}
	for _, address := range node.Status.Addresses {
		if condition := address.Type == "ExternalIP"; condition {
			filteredNode.IP = address.Address
			fmt.Printf("Node: %s with external IP: %s\n", node.ObjectMeta.Name, address.Address)
			if VerifyAvailability(dnsConfig.TargetProtocol, address.Address, dnsConfig.TargetPort, dnsConfig.TargetPath) {
				filteredNode.EndpointAvailable = true
				return filteredNode, true
			} else {
				return filteredNode, false
			}
		}
	}
	return filteredNode, false
}

func replaceNodeIfDifList(newNode NodeStatus, nodes []NodeStatus) ([]NodeStatus, bool) {
	for i, node := range nodes {
		if reflect.DeepEqual(node, newNode) {
			return nodes, false
		} else if node.IP == newNode.IP {
			nodes[i] = newNode
			return nodes, true
		}
	}
	return append(nodes, newNode), true
}