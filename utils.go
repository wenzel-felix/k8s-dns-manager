package main

import (
	"errors"
	"net"
	"strings"
)

func getDNStype(content string) (string, error) {
	if isValidIPv4Target(content) {
		return "A", nil
	} else if isValidCNAMETarget(content) {
		return "CNAME", nil
	} else {
		return "", errors.New("Could not determine DNS type.")
	}
}

func isValidIPv4Target(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	if parsedIP.To4() == nil {
		return false
	}
	return true
}

func isValidCNAMETarget(domain string) bool {
    if len(domain) < 1 || len(domain) > 255 {
        return false
    }
    if domain[len(domain)-1] == '.' {
        domain = domain[:len(domain)-1]
    }
    labels := strings.Split(domain, ".")
    for _, label := range labels {
        if len(label) < 1 || len(label) > 63 {
            return false
        }
        if label[0] == '-' || label[len(label)-1] == '-' {
            return false
        }
    }
	if _, err := net.LookupHost(domain); err != nil {
		return false
	}
    return true
}