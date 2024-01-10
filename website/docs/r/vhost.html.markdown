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
  description = "My Vhost"
  default_queue_type = "quorum"
  max_connections = "200"
  max_queues = 100
  tracing = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vhost.
* `description` - (Optional) A freindly description.
* `default_queue_type` - (Optional) default queue type for new queues
* `max_connections` - (Optional) Maximum number of concurrent client connections to the vhost
* `max_queues` - (Optional) Maximum number of queues that can be created on the vhost

## Attributes Reference

No further attributes are exported.

## Import

Vhosts can be imported using the `name`, e.g.

```
terraform import rabbitmq_vhost.my_vhost my_vhost
```
