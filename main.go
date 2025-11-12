package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/provider"
)

// value set during build by goreleaser
var (
	version = "0.1.0-alpha"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	// Allow overriding the provider address via environment variable for development
	address := os.Getenv("TF_PROVIDER_ADDRESS")
	if address == "" {
		address = "registry.terraform.io/hashicorp/contextforge"
	}

	opts := providerserver.ServeOpts{
		Address:         address,
		Debug:           debug,
		ProtocolVersion: 6,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
