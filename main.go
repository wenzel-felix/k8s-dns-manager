package main

func main() {
	defer logger.Sync()

	logger.Infow("Initialized environment variables")

	nodeChannel := make(chan NodeStatus)
	ingressChannel := make(chan Ingress)
	update := false

	go watchNodeData(nodeChannel)
	go watchIngressData(ingressChannel)

	for i := 0; ; i++ {
		select {
		case eventNode := <-nodeChannel:
			update = replaceNodeIfDifList(eventNode)
			logger.Infow("Received new node event",
				"node", eventNode.IP,
				"changes", update,
			)

		case eventIngress := <-ingressChannel:
			update = replaceIngressIfDifList(eventIngress)
			logger.Infow("Received new ingress event",
				"ingress", eventIngress.Name,
				"ingressDomains", eventIngress.Domains,
				"changes", update,
			)
		}
		if update {
			logger.Infow("Adjusting DNS entries")
			nodes, update = AdjustDNSEntries(ingresses, nodes, *dnsConfig)
		}
	}
}
