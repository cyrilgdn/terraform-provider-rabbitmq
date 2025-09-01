package rabbitmq

import (
	"context"
	"log"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcesUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcesReadUser,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourcesReadUser(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	rmqc := meta.(*rabbithole.Client)

	name := d.Get("name").(string)

	user, err := rmqc.GetUser(name)
	if err != nil {
		return diag.FromErr(checkDeleted(d, err))
	}

	log.Printf("[DEBUG] RabbitMQ: User retrieved: %#v", user)

	d.Set("name", user.Name)

	if len(user.Tags) > 0 {
		var tagList []string
		for _, v := range user.Tags {
			if v != "" {
				tagList = append(tagList, v)
			}
		}
		if len(tagList) > 0 {
			d.Set("tags", tagList)
		}
	}

	d.SetId(name)

	return diags
}
