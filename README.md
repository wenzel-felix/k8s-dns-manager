

## Required for testing
* running kubernetes cluster with a ingress installed
* kubeconfig file to connect to the cluster

## Local Docker Testing
```
# copy your kubeconfig file in the root project directory, like 
# cp ~/.kube/config kubeconfig
docker build -f DockerfileLocalKubeconfig --tag small .
docker run -e CLOUDFLARE_TOKEN="<your token>" -e CLOUDFLARE_DOMAIN="<your domain>" -e ENVIRONMENT="DEV" small
```

## Open Features
* rearrange project folder structure
* add logging
* add availability checks to targets (maybe cluster internal if possible)
* add tags support for pro plan