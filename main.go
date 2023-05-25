package main

func main() {
	defer logger.Sync()

	ingressChannel := make(chan Ingress)
	clusterUID := getUniqueClusterIdentifier()

	go watchIngressData(ingressChannel)

	for i := 0; ; i++ {
		select {
		case eventIngress := <-ingressChannel:
			adjustDNSEntries(eventIngress, clusterUID)
		}
	}
}
