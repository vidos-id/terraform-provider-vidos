package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is injected at build time via ldflags: -X main.version=<VERSION>
var version = "1.0.0"

// defaultDomain is injected at build time via ldflags: -X main.defaultDomain=<DOMAIN>
var defaultDomain = "vidos.id"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	if err := providerserver.Serve(context.Background(), New, providerserver.ServeOpts{
		Address: "registry.terraform.io/vidos/vidos",
		Debug:   debug,
	}); err != nil {
		log.Fatal(err)
	}
}
