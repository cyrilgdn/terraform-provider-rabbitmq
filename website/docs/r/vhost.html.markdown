---
layout: "rabbitmq"
page_title: "RabbitMQ: rabbitmq_vhost"
sidebar_current: "docs-rabbitmq-resource-vhost"
description: |-
  Creates and manages a vhost on a RabbitMQ server.
---

# rabbitmq\_vhost

The ``rabbitmq_vhost`` resource creates and manages a vhost.

## Example Usage

```hcl
resource "rabbitmq_vhost" "my_vhost" {
  name = "my_vhost"
  default_queue_type = "quorum"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vhost.

* `default_queue_type` - (Optional) Default queue type for newly created queues in vhost. Possible values are: `classic`, `quorum`, or `stream`. Defaults to `classic`.

## Attributes Reference

No further attributes are exported.

## Import

Vhosts can be imported using the `name`, e.g.

```
terraform import rabbitmq_vhost.my_vhost my_vhost
```
