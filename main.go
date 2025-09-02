package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-rabbitmq/rabbitmq"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debug,
		ProviderAddr: "registry.terraform.io/cyrilgdn/rabbitmq",
		ProviderFunc: rabbitmq.Provider,
	}

	plugin.Serve(opts)
}
