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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_queue_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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

	resp, err := rmqc.PutVhost(vhost, settings)
	log.Printf("[DEBUG] RabbitMQ: vhost creation response: %#v", resp)
	if err != nil {
		return err
	}

	d.SetId(vhost)

	return ReadVhost(d, meta)
}

func ReadVhost(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	vhost, err := rmqc.GetVhost(d.Id())
	if err != nil {
		return checkDeleted(d, err)
	}

	log.Printf("[DEBUG] RabbitMQ: Vhost retrieved: %#v", vhost)

	d.Set("name", vhost.Name)

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
