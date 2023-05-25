package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getUniqueClusterIdentifier returns the UID of the kube-system namespace
func getUniqueClusterIdentifier() string {
	namespace, err := kubernetesClient.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{})
	if err != nil {
		logger.Panicw("Error retrieving cluster identifier",
			"error", err,
	)	
	}
	clusterUID := string(namespace.ObjectMeta.UID)
	shortenedClusterUID := clusterUID[len(clusterUID)-12:]

	logger.Infow("Retrieved cluster identifier",
		"cluster identifier", clusterUID,
		"shortened cluster identifier", shortenedClusterUID,
	)
	return shortenedClusterUID
}
