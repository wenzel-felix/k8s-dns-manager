package main

func main() {
	defer logger.Sync()

	ingressEventChannel := make(chan Ingress)
	clusterUID := getUniqueClusterIdentifier()

	go watchIngressData(ingressEventChannel)
	go createIngressRoutine(ingressEventChannel)

	for i := 0; ; i++ {
		select {
		case eventIngress := <-ingressEventChannel:
			adjustDNSEntries(eventIngress, clusterUID)
		}
	}
}
