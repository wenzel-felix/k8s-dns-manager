package main

type DNSConfiguration struct {
	CloudflareDomain string
	CloudflareToken  string
	TargetPort       string
	TargetProtocol   string
	TargetPath       string
	IngressName      string
}

type NodeStatus struct {
	IP                string
	EndpointAvailable bool
	KubeletAvailable  bool
	DomainsConfigured bool
}

type Ingress struct {
	Name    string
	Domains []string
}
