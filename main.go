package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var sugar = logger.Sugar()

func main() {
	defer logger.Sync()

	dnsConfig := InitDNSConfiguration()
	kubernetesClient := InitKubernetesClient()

	ingresses := []Ingress{}
	nodes := []NodeStatus{}

	sugar.Infow("Initialized environment variables",
		"currentTime", time.Now(),
	)

	nodeChannel := make(chan NodeStatus)
	ingressChannel := make(chan Ingress)
	update := false

	go watchNodeData(kubernetesClient, dnsConfig, nodeChannel)
	go watchIngressData(kubernetesClient, dnsConfig, ingressChannel)

	for i := 0; ; i++ {
		select {
		case eventNode := <-nodeChannel:
			nodes, update = replaceNodeIfDifList(eventNode, nodes)
			fmt.Printf("Node: %s\n", eventNode.IP)

		case eventIngress := <-ingressChannel:
			ingresses, update = replaceIngressIfDifList(eventIngress, ingresses)
			fmt.Printf("Ingress: %s\n", eventIngress.Domains)
		}
		if update {
			fmt.Printf("Update: DNS entries\n")
			nodes, update = AdjustDNSEntries(ingresses, nodes, dnsConfig)
		}
	}
}
