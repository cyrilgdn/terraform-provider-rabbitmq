package rabbitmq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceShovelV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vhost": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"info": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ack_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "on-confirm",
						},
						"add_forward_headers": {
							Type:          schema.TypeBool,
							Optional:      true,
							ForceNew:      true,
							Default:       nil,
							ConflictsWith: []string{"info.0.destination_add_forward_headers"},
							Deprecated:    "use destination_add_forward_headers instead",
						},
						"delete_after": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       nil,
							ConflictsWith: []string{"info.0.source_delete_after"},
							Deprecated:    "use source_delete_after instead",
						},
						"destination_add_forward_headers": {
							Type:          schema.TypeBool,
							Optional:      true,
							ForceNew:      true,
							Default:       nil,
							ConflictsWith: []string{"info.0.add_forward_headers"},
						},
						"destination_add_timestamp_header": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  false,
						},
						"destination_address": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"destination_application_properties": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"destination_exchange": {
							Type:          schema.TypeString,
							ConflictsWith: []string{"info.0.destination_queue"},
							Optional:      true,
							ForceNew:      true,
							Default:       nil,
						},
						"destination_exchange_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"destination_properties": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"destination_protocol": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "amqp091",
						},
						"destination_publish_properties": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"destination_queue": {
							Type:          schema.TypeString,
							ConflictsWith: []string{"info.0.destination_exchange"},
							Default:       nil,
							Optional:      true,
							ForceNew:      true,
						},
						"destination_uri": {
							Type:      schema.TypeString,
							Required:  true,
							ForceNew:  true,
							Sensitive: false,
						},
						"prefetch_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"info.0.source_prefetch_count"},
							Deprecated:    "use source_prefetch_count instead",
							Default:       nil,
						},
						"reconnect_delay": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  1,
						},
						"source_address": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"source_delete_after": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       nil,
							ConflictsWith: []string{"info.0.delete_after"},
						},
						"source_exchange": {
							Type:          schema.TypeString,
							Default:       nil,
							ConflictsWith: []string{"info.0.source_queue"},
							Optional:      true,
							ForceNew:      true,
						},
						"source_exchange_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  nil,
						},
						"source_prefetch_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							ForceNew:      true,
							Default:       nil,
							ConflictsWith: []string{"info.0.prefetch_count"},
						},
						"source_protocol": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "amqp091",
						},
						"source_queue": {
							Type:          schema.TypeString,
							ConflictsWith: []string{"info.0.source_exchange"},
							Default:       nil,
							Optional:      true,
							ForceNew:      true,
						},
						"source_uri": {
							Type:      schema.TypeString,
							Required:  true,
							ForceNew:  true,
							Sensitive: false,
						},
					},
				},
			},
		},
	}
}

func upgradeShovelV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	propertiesUpdatedFields := []string{
		"destination_application_properties",
		"destination_properties",
		"destination_publish_properties",
	}

	if infos, ok := rawState["info"].([]any); ok {
		info := infos[0].(map[string]any)
		for _, field := range propertiesUpdatedFields {
			if v, ok := info[field].(string); ok {
				if v != "" {
					return nil, fmt.Errorf("cannot upgrade shovel when %q is set", field)
				}
				info[field] = map[string]any{}
			}
		}
		rawState["info"] = []any{info}
	}
	return rawState, nil
}
