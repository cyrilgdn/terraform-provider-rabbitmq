package rabbitmq

import (
	"context"
	"log"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcesVhost() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcesReadVhost,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourcesReadVhost(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	rmqc := meta.(*rabbithole.Client)

	name := d.Get("name").(string)

	vhost, err := rmqc.GetVhost(name)
	if err != nil {
		return diag.FromErr(checkDeleted(d, err))
	}

	log.Printf("[DEBUG] RabbitMQ: Vhost retrieved: %#v", vhost)

	d.Set("name", vhost.Name)

	d.SetId(name)

	return diags
}
