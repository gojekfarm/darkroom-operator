# Darkroom Operator

A Kubernetes Operator to put [Darkroom](https://github.com/gojek/darkroom) on Autopilot

[![Build Status](https://github.com/gojekfarm/darkroom-operator/workflows/Build/badge.svg)](https://github.com/gojekfarm/darkroom-operator/actions?query=workflow%3ABuild)
[![Test Status](https://github.com/gojekfarm/darkroom-operator/workflows/Test/badge.svg)](https://github.com/gojekfarm/darkroom-operator/actions?query=workflow%3ATest)
[![Coverage Status](https://coveralls.io/repos/github/gojekfarm/darkroom-operator/badge.svg?branch=master)](https://coveralls.io/github/gojekfarm/darkroom-operator?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gojekfarm/darkroom-operator)](https://goreportcard.com/report/github.com/gojekfarm/darkroom-operator)
[![GitHub Release](https://img.shields.io/github/release/gojekfarm/darkroom-operator.svg?style=flat)](https://github.com/gojekfarm/darkroom-operator/releases)

## Introduction

[Darkroom](https://github.com/gojek/darkroom) is a great image proxy to serve your images from your desired source and perform image manipulations on the fly.

This operator aims to make to it easy to deploy Darkroom in a Kubernetes Cluster and make it easy to manage the cluster with an intuitive GUI. 

## Installation

TBD

## Dev Environment Setup

#### Install [Operator SDK](https://sdk.operatorframework.io/docs/installation/install-operator-sdk/)

```shell script
brew install operator-sdk
```

> Note: We recommend going through operator-sdk [getting started](https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/) guide if you are new to operator development

##### Install KubeBuilder

Operator SDK uses `etcd` bundled with KubeBuilder, install it with `bin/install-kubebuilder`.
> [Optional] Add `/usr/local/kubebuilder/bin` to your $PATH

###### Creating New API Objects

The structure of this project is different than other operator codebases you might be familiar with. Since this project also includes an API Server with GUI and makes use of same API definitions, we have kept it this way.

To generate a new API object, run the command
```shell script
operator-sdk create api --group deployments --version v1alpha1 --kind Darkroom --resource=true --controller=true
```

This will generate the required resource definition and controller under `./api/` and `./controllers/`, move these to `./pkg/api/` and `./internal/controllers/` respectively. But the command will fail as there is no root level `main.go` file.

The changes required in `main.go` are hence skipped and you must add the generated resource scheme to runtime scheme in `cmd/operator/cmd/root.go` and setup the Reconciler with the Controller Manager manually.

You can run `make operator/generate` to generate necessary code by the `controller-gen`. And run `make operator/manifests` to generate the required YAML definitions.

##### Deploying Webhooks for development

It is recommended to use a [Kind](https://kind.sigs.k8s.io/) cluster for faster iteration.

Install `cert-manager` with
```shell script
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.0/cert-manager.yaml
```
> Note: You may use the latest version

### Contributing Guide

Read our [contributing guide](./CONTRIBUTING.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Darkroom Operator.

## License

Darkroom Operator is [MIT licensed](./LICENSE).
