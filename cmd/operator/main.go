package main

import "github.com/gojekfarm/darkroom-operator/cmd/operator/cmd"

func main() {
	_ = cmd.NewRootCmd().Execute()
}
