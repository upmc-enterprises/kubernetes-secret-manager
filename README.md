# Kubernetes Secret Manager

## Goals

The main motivation of this project is to allow dynamic secrets to be requested from a MySQL database and enable a pod inside a Kubernetes cluster to consume those dynamic passwords. The secrets should be tied to a lease so they expire after a pre-defined ttl and the secrets should be rotated.

We also want to do this in a way where the pod does not have to understand a specific secret generation or consumption (e.g. Hashicorp Vault).

## Usage

- Deployment Guide

## TL;DR

1. Deploy Vault (`kubectl create -f deployments/vault.yaml`)
- Configure Vault (`kubectl exec -it <podName> /bin/dumb-init /bin/sh`)
  - Run config script:  `setup-vault.sh`
- Create custom extension (`kubectl create -f thirdpartyresource/customSecret.yaml`)
- Create sample secrets (`kubectl create -f customSecret/sample-app.yaml`)
  - NOTE: This creates 2 custom secrets will in turn request two MySQL accounts from Vault, a readonly and full access account. They will be stored in Kubernetes secrets named: `db-readonly-credentials` && `db-full-credentials`



## Thanks!

Special thanks goes out to [Kelsey Hightower](https://twitter.com/kelseyhightower) for the base ideas of this project: ([https://github.com/kelseyhightower/kube-cert-manager]())

## About

Built by UPMC Enterprises in Pittsburgh, PA. http://enterprises.upmc.com/
