package main

import "github.com/gojekfarm/darkroom-operator/cmd/api-server/cmd"

func main() {
	_ = cmd.NewRootCmd().Execute()
}
