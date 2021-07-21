---
layout: "rabbitmq"
page_title: "RabbitMQ: rabbitmq_operator_policy"
sidebar_current: "docs-rabbitmq-resource-operator-policy"
description: |-
  Creates and manages an operator policy on a RabbitMQ server.
---

# rabbitmq\_operator_policy

The ``rabbitmq_operator_policy`` resource creates and manages operator policies for queues.

## Example Usage

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

resource "rabbitmq_operator_policy" "test" {
  name  = "test"
  vhost = "${rabbitmq_permissions.guest.vhost}"

  policy {
    pattern  = ".*"
    priority = 0
    apply_to = "queues"

    definition = {
      message-ttl = 3600000
      expires = 1800000
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the operator policy.

* `vhost` - (Required) The vhost to create the resource in.

* `policy` - (Required) The settings of the operator policy. The structure is
  described below.

The `policy` block supports:

* `pattern` - (Required) A pattern to match an exchange or queue name.
* `priority` - (Required) The policy with the greater priority is applied first.
* `apply_to` - (Required) Can be "queues".
* `definition` - (Required) Key/value pairs of the operator policy definition. See the
  RabbitMQ documentation for definition references and examples.

## Attributes Reference

No further attributes are exported.

## Import

Operator policies can be imported using the `id` which is composed of `name@vhost`.
E.g.

```
terraform import rabbitmq_operator_policy.test name@vhost
```
