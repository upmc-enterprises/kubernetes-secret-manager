# Deployment Guide

This guide will walk you through deploying the Kubernetes Secret Manager.

## Deploying

### MySQL

MySQL is the backend used to demonstrate how to get user accounts dynamically from Vault.

```
kubectl create -f deployments/mysql.yaml
```

### Vault

This project uses Vault from Hashicorp as it's secret backend to dynamically create secrets. A sample Vault deployment is included which should only be utilized as a proof of concept, it is not intended for a production use.

```
kubectl create -f deployments/vault.yaml
```

The sample Vault instance included in this project will be running in developer mode which will not require unsealing, however, it will still need to be configured with some policies we'll use later in the demo.

```
kubectl exec -it <podName> /bin/dumb-init /bin/sh
> setup-vault.sh
```

#### Vault Configuration

The `setup-vault.sh` script creates some default policies which are configured in the file [myapp.hcl](deployments/vault/myapp.hcl).

### Custom ThirdPartyResource

A custom [ThirdPartyResource](https://github.com/kubernetes/kubernetes/blob/release-1.3/docs/design/extending-api.md) is required to be created by an application.

```
kubectl create -f thirdpartyresource/customSecret.yaml
```
Once the ThirdPartyResource is created you can create the custom secret object which utilized this new resource:

```
kubectl create -f customSecret/sample-app.yaml
```

In the sample app yaml file outlines the following config parametes:
- secret: Name of the secret to create in Kubernetes
- policy: Policy to request from Vault
