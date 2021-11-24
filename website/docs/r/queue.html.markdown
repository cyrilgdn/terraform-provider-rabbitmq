---
layout: "rabbitmq"
page_title: "RabbitMQ: rabbitmq_queue"
sidebar_current: "docs-rabbitmq-resource-queue"
description: |-
  Creates and manages a queue on a RabbitMQ server.
---

# rabbitmq\_queue

The ``rabbitmq_queue`` resource creates and manages a queue.

## Example Usage

### Basic Example

```hcl
resource "rabbitmq_vhost" "test" {
  name = "test"
}

resource "rabbitmq_permissions" "guest" {
  user  = "guest"
  vhost = "${rabbitmq_vhost.test.name}"

  permissions {
    configure = ".*"
    write     = ".*"
    read      = ".*"
  }
}

resource "rabbitmq_queue" "test" {
  name  = "test"
  vhost = "${rabbitmq_permissions.guest.vhost}"

  settings {
    durable     = false
    auto_delete = true
    arguments = {
      "x-queue-type" : "quorum",
    }
  }
}
```

### Example With JSON Arguments

```hcl
variable "arguments" {
  default = <<EOF
{
  "x-message-ttl": 5000
}
EOF
}

resource "rabbitmq_vhost" "test" {
  name = "test"
}

resource "rabbitmq_permissions" "guest" {
  user  = "guest"
  vhost = "${rabbitmq_vhost.test.name}"

  permissions {
    configure = ".*"
    write     = ".*"
    read      = ".*"
  }
}

resource "rabbitmq_queue" "test" {
  name  = "test"
  vhost = "${rabbitmq_permissions.guest.vhost}"

  settings {
    durable     = false
    auto_delete = true
    arguments_json = "${var.arguments}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the queue.

* `vhost` - (Required) The vhost to create the resource in.

* `settings` - (Required) The settings of the queue. The structure is
  described below.

The `settings` block supports:

* `durable` - (Optional) Whether the queue survives server restarts.
  Defaults to `false`.

* `auto_delete` - (Optional) Whether the queue will self-delete when all
  consumers have unsubscribed.

* `arguments` - (Optional) Additional key/value settings for the queue.
  All values will be sent to RabbitMQ as a string. If you require non-string
  values, use `arguments_json`.

* `arguments_json` - (Optional) A nested JSON string which contains additional
  settings for the queue. This is useful for when the arguments contain
  non-string values.

## Attributes Reference

* `x-queue-type` - (Optional) this argument sets the queue type. Possible values: `classic` (default), `quorum`, `stream` .

No further attributes are exported.

## Import

Queues can be imported using the `id` which is composed of `name@vhost`. E.g.

```
terraform import rabbitmq_queue.test name@vhost
```
