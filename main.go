package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var logger, _ = zap.NewProduction()
var sugar = logger.Sugar()

func main() {
	defer logger.Sync()

	serverPort := getEnvOrDefault("TARGET_PORT", "80")
	protocol := getEnvOrDefault("TARGET_PROTOCOL", "http")
	path := getEnvOrDefault("TARGET_PATH", "/healthz")
	cloudflareDomain := getEnvOrDefault("CLOUDFLARE_DOMAIN", "")
	cloudflareToken := getEnvOrDefault("CLOUDFLARE_TOKEN", "")
	ingressName := getEnvOrDefault("INGRESS_NAME", "")

	sugar.Infow("Initialized environment variables",
		"currentTime", time.Now(),
	)

	domains, targets := fetchNodeData(ingressName)

	var hostStatuses []Host

	for _, host := range targets {
		if condition := isValidIPv4Address(host); condition {
			status := verifyAvailability(protocol, host, serverPort, path)
			hostStatuses = append(hostStatuses, Host{IP: host, Available: status})
		} else {
			fmt.Printf("Skipping host: %s as it is not a valid IPv4 address.\n", host)
		}
	}

	api, err := cloudflare.NewWithAPIToken(cloudflareToken)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(cloudflareDomain)
	if err != nil {
		log.Fatal(err)
	}

	for _, host := range hostStatuses {
		for _, domain := range domains {
			fmt.Printf("host: %s, available: %t\n", host.IP, host.Available)
			recs, info, err := api.ListDNSRecords(context.Background(),
				cloudflare.ZoneIdentifier(zoneID),
				cloudflare.ListDNSRecordsParams{Type: "A", Content: host.IP, Name: domain})

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Existing records: %+v\n", info.Count)

			if condition := (host.Available && info.Count == 0); condition {
				record, err := api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
					Type:    "A",
					Name:    domain,
					Content: host.IP,
				})
				if err != nil {
					log.Println(err)
				} else {
					fmt.Printf("A record created: %s => %s\n", record.Name, record.Content)
				}
			} else if condition := (!host.Available && info.Count > 0); condition {
				for _, record := range recs {
					err := api.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), record.ID)
					if err != nil {
						log.Println(err)
					} else {
						fmt.Printf("A record deleted: %s => %s\n", record.Name, record.Content)
					}
				}
			} else if condition := (host.Available && info.Count > 0); condition {
				fmt.Printf("A record already exists: %s => %s\n", domain, host.IP)
			}
		}
	}
}

func fetchNodeData(ingressName string) ([]string, []string) {

	config := &rest.Config{}
	initErr := error(nil)

	if condition := getEnvOrDefault("ENVIRONMENT", "PRD") == "DEV"; condition {
		fmt.Printf("client: using local kubeconfig\n")
		config, initErr = clientcmd.BuildConfigFromFlags("", "kubeconfig")
		if initErr != nil {
			panic(initErr.Error())
		}
	} else {
		fmt.Printf("client: using in-cluster kubeconfig\n")
		config, initErr = rest.InClusterConfig()
		if initErr != nil {
			panic(initErr.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	ingressList, err := clientset.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{FieldSelector: "metadata.name=" + ingressName})
	hostList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}
	sugar.Infow("Initialized environment variables",
		"currentTime", time.Now(),
	)

	domains := []string{}
	targets := []string{}

	for _, ingress := range ingressList.Items {
		if condition := ingress.ObjectMeta.Name == ingressName; condition {
			for _, rule := range ingress.Spec.TLS {
				for _, host := range rule.Hosts {
					domains = append(domains, host)
				}
			}
		}
	}

	for _, host := range hostList.Items {
		addToTargets := true
		nodeConditions := host.Status.Conditions
		for _, condition := range nodeConditions {
			if condition.Reason == "KubeletReady" {
				if condition.Type == "Ready" && condition.Status == "True" {
					fmt.Printf("Found ready node: %s\n", host.ObjectMeta.Name)
				} else {
					addToTargets = false
				}
			}
		}

		if condition := addToTargets; condition {
			fmt.Printf("Targeting node: %s\n", host.ObjectMeta.Name)
			for _, address := range host.Status.Addresses {
				if condition := address.Type == "ExternalIP"; condition {
					targets = append(targets, address.Address)
				}
			}
		}
	}

	fmt.Printf("Found domains: %+v\n", domains)
	fmt.Printf("Found targets: %+v\n", targets)
	return domains, targets
}

func verifyAvailability(protocol string, host string, serverPort string, path string) bool {
	requestURL := fmt.Sprintf("%s://%s:%s%s", protocol, host, serverPort, path)

	fmt.Printf("client: making request to %s\n", requestURL)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return false
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return false
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	if condition := res.StatusCode == http.StatusOK; !condition {
		fmt.Printf("client: status code not OK: %d\n", res.StatusCode)
		return false
	}

	return true
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			fmt.Printf("No value for required variable %s provided\n", key)
			os.Exit(1)
		}
		return defaultValue
	}
	return value
}

func isValidIPv4Address(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	if parsedIP.To4() == nil {
		return false
	}
	return true
}

type Host struct {
	IP        string
	Available bool
}

type IngressList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name            string            `json:"name"`
			Namespace       string            `json:"namespace"`
			UID             string            `json:"uid"`
			ResourceVersion string            `json:"resourceVersion"`
			Generation      int               `json:"generation"`
			CreationTime    string            `json:"creationTimestamp"`
			Annotations     map[string]string `json:"annotations"`
			ManagedFields   []struct {
				Manager    string `json:"manager"`
				Operation  string `json:"operation"`
				APIVersion string `json:"apiVersion"`
				Time       string `json:"time"`
				FieldsType string `json:"fieldsType"`
				FieldsV1   struct {
					Metadata struct {
						Annotations struct {
							CertManagerClusterIssuer string `json:"cert-manager.io/cluster-issuer"`
						} `json:"annotations"`
					} `json:"metadata"`
					Spec struct {
						IngressClassName string `json:"ingressClassName"`
						Rules            []struct {
							Host string `json:"host"`
							HTTP struct {
								Paths []struct {
									Path     string `json:"path"`
									PathType string `json:"pathType"`
									Backend  struct {
										Service struct {
											Name string `json:"name"`
											Port struct {
												Number int `json:"number"`
											} `json:"port"`
										} `json:"service"`
									} `json:"backend"`
								} `json:"paths"`
							} `json:"http"`
						} `json:"rules"`
						TLS []struct {
							Hosts      []string `json:"hosts"`
							SecretName string   `json:"secretName"`
						} `json:"tls"`
					} `json:"spec"`
				} `json:"fieldsV1"`
				Subresource string `json:"subresource,omitempty"`
			} `json:"managedFields"`
		} `json:"metadata"`
		Spec struct {
			IngressClassName string `json:"ingressClassName"`
			TLS              []struct {
				Hosts      []string `json:"hosts"`
				SecretName string   `json:"secretName"`
			} `json:"tls"`
			Rules []struct {
				Host string `json:"host"`
				HTTP struct {
					Paths []struct {
						Path     string `json:"path"`
						PathType string `json:"pathType"`
						Backend  struct {
							Service struct {
								Name string `json:"name"`
								Port struct {
									Number int `json:"number"`
								} `json:"port"`
							} `json:"service"`
						} `json:"backend"`
					} `json:"paths"`
				} `json:"http"`
			} `json:"rules"`
		} `json:"spec"`
		Status struct {
			LoadBalancer struct {
				Ingress []struct {
					IP string `json:"ip"`
				} `json:"ingress"`
			} `json:"loadBalancer"`
		} `json:"status"`
	} `json:"items"`
}
