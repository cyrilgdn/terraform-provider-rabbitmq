package rabbitmq

import (
	"fmt"
	"log"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVhost() *schema.Resource {
	return &schema.Resource{
		Create: CreateVhost,
		Read:   ReadVhost,
		Delete: DeleteVhost,
		Update: UpdateVhost,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			// "tags": {
			// 	Type:     schema.TypeList,
			// 	Elem:     &schema.Schema{Type: schema.TypeString},
			// 	Optional: true,
			// 	ForceNew: true,
			// },
			"default_queue_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"tracing": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"max_connections": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
			"max_queues": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
		},
	}
}

func CreateVhost(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	vhost := d.Get("name").(string)

	log.Printf("[DEBUG] RabbitMQ: Attempting to create vhost %s", vhost)

	var settings rabbithole.VhostSettings

	if v, ok := d.Get("default_queue_type").(string); ok && v != "" {
		settings.DefaultQueueType = v
	}

	if v, ok := d.Get("description").(string); ok && v != "" {
		settings.Description = v
	}

	if v, ok := d.Get("tracing").(bool); ok {
		settings.Tracing = v
	}

	limits := make(rabbithole.VhostLimitsValues)

	if v, ok := d.GetOk("max_connections"); ok {
		limits["max-connections"] = v.(int)
	}

	if v, ok := d.GetOk("max_queues"); ok {
		limits["max-queues"] = v.(int)
	}

	resp, err := rmqc.PutVhost(vhost, settings)
	log.Printf("[DEBUG] RabbitMQ: vhost creation response: %#v", resp)
	if err != nil {
		return err
	}

	resp, err = rmqc.PutVhostLimits(vhost, limits)
	log.Printf("[DEBUG] RabbitMQ: vhost creation response: %#v", resp)
	if err != nil {
		return err
	}

	d.SetId(vhost)

	return ReadVhost(d, meta)
}

func UpdateVhost(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	vhost, err := rmqc.GetVhost(d.Id())
	if err != nil {
		return checkDeleted(d, err)
	}

	vhost_limits_info, err := rmqc.GetVhostLimits(d.Id())
	if err != nil {
		return checkDeleted(d, err)
	}

	var settings rabbithole.VhostSettings
	limits := make(rabbithole.VhostLimitsValues)

	if d.HasChange("description") {
		_, newDescription := d.GetChange("description")

		if v, ok := newDescription.(string); ok && v != "" {
			settings.Description = v
		}
	} else {
		settings.Description = vhost.Description
	}

	if d.HasChange("default_queue_type") {
		_, newDefaultQueueType := d.GetChange("default_queue_type")

		if v, ok := newDefaultQueueType.(string); ok && v != "" {
			settings.DefaultQueueType = v
		}
	} else {
		settings.DefaultQueueType = vhost.DefaultQueueType
	}

	if d.HasChange("tracing") {
		_, newTracing := d.GetChange("tracing")

		if v, ok := newTracing.(bool); ok {
			settings.Tracing = v
		}
	} else {
		settings.Tracing = vhost.Tracing
	}

	if _, ok := d.GetOk("max_connections"); ok {
		if d.HasChange("max_connections") {
			_, newMaxConnections := d.GetChange("max_connections")

			if v, ok := newMaxConnections.(int); ok {
				limits["max-connections"] = v
			}
		} else {
			limits["max-connections"] = vhost_limits_info[0].Value["max-connections"]
		}
	}

	if _, ok := d.GetOk("max_queues"); ok {
		if d.HasChange("max_queues") {
			_, newMaxQueues := d.GetChange("max_queues")

			if v, ok := newMaxQueues.(int); ok {
				limits["max-queues"] = v
			}
		} else {
			limits["max-queues"] = vhost_limits_info[0].Value["max-queues"]
		}
	}

	resp, err := rmqc.PutVhost(vhost.Name, settings)
	log.Printf("[DEBUG] RabbitMQ: vhost creation response: %#v", resp)
	if err != nil {
		return err
	}

	resp, err = rmqc.DeleteVhostLimits(vhost.Name, rabbithole.VhostLimits{"max-connections", "max-queues"})
	log.Printf("[DEBUG] RabbitMQ: vhost limits deletion response: %#v", resp)
	if err != nil {
		return err
	}

	resp, err = rmqc.PutVhostLimits(vhost.Name, limits)
	log.Printf("[DEBUG] RabbitMQ: vhost limits creation response: %#v", resp)
	if err != nil {
		return err
	}

	return ReadVhost(d, meta)
}

func ReadVhost(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	vhost, err := rmqc.GetVhost(d.Id())
	if err != nil {
		return checkDeleted(d, err)
	}

	log.Printf("[DEBUG] RabbitMQ: Vhost retrieved: %#v", vhost)

	vhost_limits_info, err := rmqc.GetVhostLimits(d.Id())
	if err != nil {
		return checkDeleted(d, err)
	}

	log.Printf("[DEBUG] RabbitMQ: Vhost retrieved: %#v", vhost)

	d.Set("name", vhost.Name)
	d.Set("description", vhost.Description)
	// d.Set("tags", vhost.Tags)
	d.Set("tracing", vhost.Tracing)

	if len(vhost_limits_info) > 0 {
		if val, ok := vhost_limits_info[0].Value["max-connections"]; ok {
			if val > -1 {
				d.Set("max_connections", val)
			}
		}

		if val, ok := vhost_limits_info[0].Value["max-queues"]; ok {
			if val > -1 {
				d.Set("max_queues", val)
			}
		}

	}

	return nil
}

func DeleteVhost(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	log.Printf("[DEBUG] RabbitMQ: Attempting to delete vhost %s", d.Id())

	resp, err := rmqc.DeleteVhost(d.Id())
	log.Printf("[DEBUG] RabbitMQ: vhost deletion response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		// the vhost was automatically deleted
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error deleting RabbitMQ user: %s", resp.Status)
	}

	return nil
}
