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

### Contributing Guide

Read our [contributing guide](./CONTRIBUTING.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Darkroom Operator.

## License

Darkroom Operator is [MIT licensed](./LICENSE).