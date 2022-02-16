## Sidecar Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/ChinaLHR/sidecar-operator/blob/main/LICENSE)
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/chinalhr/sidecar-operator)](https://img.shields.io/github/go-mod/go-version/chinalhr/sidecar-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/ChinaLHR/sidecar-operator)](https://goreportcard.com/report/github.com/ChinaLHR/sidecar-operator)
[![CircleCI](https://circleci.com/gh/ChinaLHR/sidecar-operator/tree/main.svg?style=shield)](https://circleci.com/gh/ChinaLHR/sidecar-operator/tree/main)

## Overview

sidecar-operator manages sidecar in the kubernetes cluster and injects sidecar into Pod in the cluster. It is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk).

## Quick Start

### install Operator SDK

See Operator SDK document：

[https://sdk.operatorframework.io/docs/installation/](https://sdk.operatorframework.io/docs/installation/)

### Deploying the cert manager in the kubernetes cluster

We using [cert manager](https://github.com/jetstack/cert-manager) for provisioning the certificates for the webhook server.

See cert manager document:

[https://cert-manager.io/docs/installation/](https://cert-manager.io/docs/installation/)

### Deploying the Sidecar Operator in the kubernetes cluster

Build and Push the docker image to your repository.

deploying the sidecar operator to your kubernetes cluster.

```bash
make docker-build docker-push IMG=*/sidecar-operator:1.0.0
make deploy IMG=*/sidecar-operator:1.0.0
```

### Create a simple SidecarSet

See the yaml file in the /example/config directory.