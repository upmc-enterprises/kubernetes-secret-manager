# Kubernetes Secret Manager

## Problem

Typically usernames and passwords to resources are statically tied to a service account. These passwords rarely change and are usually difficult to rotate in an application stack. Sometimes, we're not even sure how many components are utilizing that service account which makes rotate even more difficult and teams end up not changing due to fear of downtime and errors.

Ideally we want a solution which allows us to rotate credentials dynamically and do so in a secure well-thought out way.

## Goals

The main motivation of this project is to allow dynamic secrets to be requested from a MySQL database and enable a pod inside a Kubernetes cluster to consume those dynamic passwords. The secrets should be tied to a lease so they expire after a pre-defined ttl and the secrets should be rotated before a max ttl is met.

The implementation should be done so that the pod does not have to understand a specific secret generation tool (e.g. Hashicorp Vault). The application only needs to understand how to read from a file as well as get notified when that file changes.

## Implementation

This project uses [Vault](https://www.vaultproject.io/) as it's secret distibution tool with the [MySQL Secret Backend](https://www.vaultproject.io/docs/secrets/mysql/index.html) enabled. It's deployed via a custom `ThirdPartyResource` and kubernetes controller which implements the Vault API. Credentials are exposed to pods via simple Kubernetes secrets. The application in the pod is only responsible for refreshing it's application state when those credentials are rotated.

#### Video Walkthrough
[![Kubernetes Secret Manager](http://img.youtube.com/vi/kb7DU-Qwtrc/0.jpg)](http://www.youtube.com/watch?v=kb7DU-Qwtrc)

## Usage

- [Deployment Guide](docs/deployment-guide.md)

## TL;DR

1. Deploy mysql (`kubectl create -f deployments/mysql.yaml`) 
- Deploy Vault (`kubectl create -f deployments/vault.yaml`)
- Configure Vault (`kubectl exec -it <podName> /bin/dumb-init /bin/sh`)
  - Run config script:  `setup-vault.sh`
- Create custom extension (`kubectl create -f thirdpartyresource/customSecret.yaml`)
- Create sample app (`kubectl create -f sample-app/deployments/sample-app.yaml`)
  - NOTE: This creates 2 custom secrets will in turn request two MySQL accounts from Vault, a readonly and full access account. They will be stored in Kubernetes secrets named: `db-readonly-credentials` && `db-full-credentials`



## Thanks!

Special thanks goes out to [Kelsey Hightower](https://twitter.com/kelseyhightower) for the base ideas of this project: ([https://github.com/kelseyhightower/kube-cert-manager]())

## About

Built by UPMC Enterprises in Pittsburgh, PA. http://enterprises.upmc.com/
