---
layout: "rabbitmq"
page_title: "Provider: RabbitMQ"
sidebar_current: "docs-rabbitmq-index"
description: |-
  A provider for a RabbitMQ Server.
---

# RabbitMQ Provider

[RabbitMQ](http://www.rabbitmq.com) is an AMQP message broker server. The
RabbitMQ provider exposes resources used to manage the configuration of
resources in a RabbitMQ server.

Use the navigation to the left to read about the available resources.

## Example Usage

The following is a minimal example:

```hcl
# Configure the RabbitMQ provider
provider "rabbitmq" {
  endpoint = "http://127.0.0.1"
  username = "guest"
  password = "guest"
}

# Create a virtual host
resource "rabbitmq_vhost" "vhost_1" {
  name = "vhost_1"
}
```

## Requirements

The RabbitMQ management plugin must be enabled on the server, to use this provider. You can
enable the plugin by doing something similar to:

```
$ sudo rabbitmq-plugins enable rabbitmq_management
```

## Argument Reference

The following arguments are supported:

* `endpoint` - (Required) The HTTP URL of the management plugin on the
  RabbitMQ server. This can also be sourced from the `RABBITMQ_ENDPOINT`
  Environment Variable. The RabbitMQ management plugin *must* be enabled in order
  to use this provider. _Note_: This is not the IP address or hostname of the
  RabbitMQ server that you would use to access RabbitMQ directly.
* `username` - (Required) Username to use to authenticate with the server.
  This can also be sourced from the `RABBITMQ_USERNAME` Environment Variable.
* `password` - (Optional) Password for the given user. This can also be sourced
  from the `RABBITMQ_PASSWORD` Environment Variable.
* `insecure` - (Optional) Boolean. Trust self-signed certificates. This can also be sourced
  from the `RABBITMQ_INSECURE` Environment Variable.
* `cacert_file` - (Optional) The path to a custom CA / intermediate certificate.
  This can also be sourced from the `RABBITMQ_CACERT` Environment Variable.
* `clientcert_file` - (Optional) The path to the X.509 client certificate.
  This can also be sourced from the `RABBITMQ_CLIENTCERT` Environment Variable
* `clientkey_file` - (Optional) The path to the private key.
  This can also be sourced from the `RABBITMQ_CLIENTKEY` Environment Variable
* `proxy` - (Optional) The URL of a proxy through which to send HTTP requests to
  the RabbitMQ server. This can also be sourced from the `RABBITMQ_PROXY`
  Environment Variable. If not set, the default `HTTP_PROXY`/`HTTPS_PROXY` will
  be used instead.
