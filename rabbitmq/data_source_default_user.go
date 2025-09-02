package rabbitmq

import (
	"context"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcesDefaultUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcesReadDefaultUser,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourcesReadDefaultUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	rmqc := meta.(*rabbithole.Client)

	username := rmqc.Username

	d.Set("username", username)
	d.SetId(username)
	return diags
}
