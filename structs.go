package main

type DNSConfiguration struct {
	CloudflareDomain string
	CloudflareToken  string
	TargetPort       string
	TargetProtocol   string
	TargetPath       string
	IngressName      string
}

type Ingress struct {
	Name    string
	Domains []string
	Targets []string
}
