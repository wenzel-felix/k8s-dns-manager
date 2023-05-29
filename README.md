<div align="center" width="100%">
    <h2>hcloud rke2 module</h2>
    <p>Simple and fast creation of a rke2 Kubernetes cluster on Hetzner Cloud.</p>
    <a target="_blank" href="https://github.com/wenzel-felix/terraform-hcloud-rke2/stargazers"><img src="https://img.shields.io/github/stars/wenzel-felix/terraform-hcloud-rke2" /></a>
    <a target="_blank" href="https://github.com/wenzel-felix/terraform-hcloud-rke2/releases"><img src="https://img.shields.io/github/v/release/wenzel-felix/terraform-hcloud-rke2?display_name=tag" /></a>
    <a target="_blank" href="https://github.com/wenzel-felix/terraform-hcloud-rke2/commits/master"><img src="https://img.shields.io/github/last-commit/wenzel-felix/terraform-hcloud-rke2" /></a>
</div>

## ✨ Features

- Manage your DNS records for your ingress resources automatically
- Fast and easy to use
- Deploy to you own cluster via Helm

## 🤔 Why?

This application provides a Kubernetes-native solution for managing DNS entries based on ingresses. It watches for changes to ingresses and automatically updates DNS records to reflect the current state of the cluster.

## 🔧 Prerequisites

There are no special prerequirements in order to take advantage of this module. Only things required are:
* Domain hosted on Cloudflare & a Cloudflare API Token to manage it

## 🚀 Usage

### Get Repository Info

```console
helm repo add k8s-dns-manager https://wenzel-felix.github.io/k8s-dns-manager/
helm repo update
```

_See [`helm repo`](https://helm.sh/docs/helm/helm_repo/) for command documentation._

### Install Chart

```console
helm install [RELEASE_NAME] k8s-dns-manager/k8s-dns-manager
```

_See [configuration](#configuring) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

### Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

### Upgrading Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._