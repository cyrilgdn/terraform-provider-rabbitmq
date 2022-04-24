---
layout: "rabbitmq"
page_title: "RabbitMQ: rabbitmq_limit"
sidebar_current: "docs-rabbitmq-resource-limit"
description: |-
Creates and manages user or vhost limit on a RabbitMQ server.
---

# rabbitmq\_limit

The ``rabbitmq_limit`` resource creates and manages a limit.

## Example Usage

```hcl
resource "rabbitmq_limit" "my_user_limit" {

	scope = "user"
	alias = "guest"
	limit = "max-channels"
	value = 200
 }
```

```hcl
resource "rabbitmq_limit" "my_vhost_limit" {

	scope = "vhost"
	alias = "/"
	limit = "max-connections"
	value = 200
 }
```

## Argument Reference

The following arguments are supported:

* `scope` - (Required) Limit scope: user or vhost.
* `alias` - (Required) Limit name: max-connection, max-channel.
* `limit` - (Required) Limit alias: username or vhost name.
* `value` - (Required) Limit value.

## Attributes Reference

No further attributes are exported.

## Import

Limit can be imported using the `id` which is composed of `scope@limit@alias`. E.g.

```
terraform import rabbitmq_limit.my_user_limit  user@max-channels@guest
```
