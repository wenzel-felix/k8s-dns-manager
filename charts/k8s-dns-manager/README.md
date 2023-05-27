# k8s-dns-manager

This Helm chart provides a Kubernetes-native solution for managing DNS entries based on ingresses. The chart includes a Kubernetes controller that watches for changes to ingresses and automatically updates DNS records to reflect the current state of the cluster. 

## Get Repository Info

```console
helm repo add k8s-dns-manager https://wenzel-felix.github.io/k8s-dns-manager/
helm repo update
```

_See [`helm repo`](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Chart

```console
helm install [RELEASE_NAME] k8s-dns-manager/k8s-dns-manager
```

_See [configuration](#configuring) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._