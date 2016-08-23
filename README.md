# Kubernetes Vault Manager

## Vault

1. Deploy Vault (`kubectl create -f deployments/vault.yaml`)
- Configure Vault (`kubectl exec -it <podName> /bin/dumb-init /bin/sh`)
- Run config script:  `setup-vault.sh`
