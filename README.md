

## Required for testing
* running kubernetes cluster with a ingress installed
* kubeconfig file to connect to the cluster

## Local Docker Testing
```
# copy your kubeconfig file in the root project directory, like 
# cp ~/.kube/config kubeconfig
docker build -f DockerfileLocalKubeconfig --tag small .
docker run -e CLOUDFLARE_TOKEN="<your token>" -e CLOUDFLARE_DOMAIN="<your domain>" -e INGRESS_NAME="<name of your ingress>" -e ENVIRONMENT="DEV" small
```

## Open Features
* Use CloudFlare Tags to check if node is completely gone