package main

func main() {
	defer logger.Sync()

	ingressChannel := make(chan Ingress)
	update := false
	clusterUID := getUniqueClusterIdentifier()

	go watchIngressData(ingressChannel)

	for i := 0; ; i++ {
		select {
		case eventIngress := <-ingressChannel:
			logger.Infow("Received new ingress event",
				"ingress", eventIngress.Name,
				"ingressDomains", eventIngress.Domains,
				"changes", update,
			)
			adjustDNSEntries(eventIngress, clusterUID)
		}
	}
}
