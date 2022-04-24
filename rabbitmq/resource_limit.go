package rabbitmq

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func resourceLimit() *schema.Resource {

	return &schema.Resource{

		Create: CreateLimit,
		Read:   ReadLimit,
		Update: CreateLimit,
		Delete: DeleteLimit,

		Importer: &schema.ResourceImporter{

			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			"scope": {

				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Limit scope: user or vhost",
				ValidateFunc: validation.StringInSlice([]string{"user", "vhost"}, true),
			},

			"limit": {

				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Limit name: max-connection, max-channel",
			},

			"alias": {

				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Limit alias: username or vhost name",
			},

			"value": {

				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func CreateLimit(d *schema.ResourceData, meta interface{}) error {

	scope := d.Get("scope").(string)
	limit := d.Get("limit").(string)
	alias := d.Get("alias").(string)
	value := d.Get("value").(int)

	log.Printf("[DEBUG] RabbitMQ: Attempting to create %s limit for %s with value %d", limit, alias, value)

	rmqc := meta.(*rabbithole.Client)

	var (
		resp *http.Response
		err  error
	)

	switch scope {

	case "user":
		resp, err = rmqc.PutUserLimits(alias, map[string]int{limit: value})
	case "vhost":
		resp, err = rmqc.PutVhostLimits(alias, map[string]int{limit: value})
	}

	log.Printf("[DEBUG] RabbitMQ: limit creation response: %#v", resp)

	if err != nil {

		return err
	}

	d.SetId(fmt.Sprintf("%s@%s@%s", scope, limit, alias))

	return ReadLimit(d, meta)
}

func ReadLimit(d *schema.ResourceData, meta interface{}) error {

	scope, limit, alias, err := parseLimitID(d.Id())

	if err != nil {

		return err
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to retrieve limits for %s", alias)

	rmqc := meta.(*rabbithole.Client)

	var limits map[string]int

	switch scope {

	case "user":
		limits, err = getLimits(rmqc.GetUserLimits(alias))
	case "vhost":
		limits, err = getLimits(rmqc.GetVhostLimits(alias))
	}

	if err != nil {

		return err
	}

	log.Printf("[DEBUG] RabbitMQ: Limit %s retrieved for %s", limit, alias)

	if value, ok := limits[limit]; ok {

		_ = d.Set("scope", scope)
		_ = d.Set("alias", alias)
		_ = d.Set("value", value)

		log.Printf("[DEBUG] RabbitMQ: Limit value %d:", value)

		return nil
	}

	return fmt.Errorf("limit %s was not found", limit)
}

func DeleteLimit(d *schema.ResourceData, meta interface{}) error {

	scope, limit, alias, err := parseLimitID(d.Id())

	if err != nil {

		return err
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to delete limit %s for %s", limit, alias)

	rmqc := meta.(*rabbithole.Client)

	var resp *http.Response

	switch scope {

	case "user":
		resp, err = rmqc.DeleteUserLimits(alias, []string{limit})
	case "vhost":
		resp, err = rmqc.DeleteVhostLimits(alias, []string{limit})
	}

	log.Printf("[DEBUG] RabbitMQ: limit deletion response: %#v", resp)

	return err
}
