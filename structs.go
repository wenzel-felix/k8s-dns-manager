package main

type DNSConfiguration struct {
	CloudflareDomain string
	CloudflareToken  string
	TargetPort       string
	TargetProtocol   string
	TargetPath       string
}

type Ingress struct {
	Name    string
	Domains []string
	Targets []string
}
