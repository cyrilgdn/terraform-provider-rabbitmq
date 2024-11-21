package rabbitmq

import (
	"fmt"
	"log"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceShovel() *schema.Resource {
	return &schema.Resource{
		Create: CreateShovel,
		Update: UpdateShovel,
		Read:   ReadShovel,
		Delete: DeleteShovel,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
							Type:     schema.TypeMap,
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
							Type:     schema.TypeMap,
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
							Type:     schema.TypeMap,
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

func CreateShovel(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	vhost := d.Get("vhost").(string)
	shovelName := d.Get("name").(string)
	shovelInfo := d.Get("info").([]interface{})

	shovelMap, ok := shovelInfo[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Unable to parse shovel info")
	}

	shovelDefinition := setShovelDefinition(shovelMap).(rabbithole.ShovelDefinition)

	log.Printf("[DEBUG] RabbitMQ: Attempting to declare shovel %s in vhost %s", shovelName, vhost)
	resp, err := rmqc.DeclareShovel(vhost, shovelName, shovelDefinition)
	log.Printf("[DEBUG] RabbitMQ: shovel declartion response: %#v", resp)
	if err != nil {
		return err
	}

	shovelId := fmt.Sprintf("%s@%s", shovelName, vhost)

	d.SetId(shovelId)

	return ReadShovel(d, meta)
}

func ReadShovel(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	shovelInfo, err := rmqc.GetShovel(vhost, name)
	if err != nil {
		return checkDeleted(d, err)
	}

	log.Printf("[DEBUG] RabbitMQ: Shovel retrieved: Vhost: %#v, Name: %#v", vhost, name)

	info := make(map[string]interface{})
	info["ack_mode"] = shovelInfo.Definition.AckMode
	info["add_forward_headers"] = shovelInfo.Definition.AddForwardHeaders
	info["delete_after"] = shovelInfo.Definition.DeleteAfter
	info["destination_add_forward_headers"] = shovelInfo.Definition.DestinationAddForwardHeaders
	info["destination_add_timestamp_header"] = shovelInfo.Definition.DestinationAddTimestampHeader
	info["destination_address"] = shovelInfo.Definition.DestinationAddress
	info["destination_application_properties"] = shovelInfo.Definition.DestinationApplicationProperties
	info["destination_exchange"] = shovelInfo.Definition.DestinationExchange
	info["destination_exchange_key"] = shovelInfo.Definition.DestinationExchangeKey
	info["destination_properties"] = shovelInfo.Definition.DestinationProperties
	info["destination_protocol"] = shovelInfo.Definition.DestinationProtocol
	info["destination_publish_properties"] = shovelInfo.Definition.DestinationPublishProperties
	info["destination_queue"] = shovelInfo.Definition.DestinationQueue
	if len(shovelInfo.Definition.DestinationURI) > 0 {
		info["destination_uri"] = shovelInfo.Definition.DestinationURI[0]
	}
	info["prefetch_count"] = shovelInfo.Definition.PrefetchCount
	info["reconnect_delay"] = shovelInfo.Definition.ReconnectDelay
	info["source_address"] = shovelInfo.Definition.SourceAddress
	info["source_delete_after"] = shovelInfo.Definition.SourceDeleteAfter
	info["source_exchange"] = shovelInfo.Definition.SourceExchange
	info["source_exchange_key"] = shovelInfo.Definition.SourceExchangeKey
	info["source_prefetch_count"] = shovelInfo.Definition.SourcePrefetchCount
	info["source_protocol"] = shovelInfo.Definition.SourceProtocol
	info["source_queue"] = shovelInfo.Definition.SourceQueue
	if len(shovelInfo.Definition.SourceURI) > 0 {
		info["source_uri"] = shovelInfo.Definition.SourceURI[0]
	}

	d.Set("name", shovelInfo.Name)
	d.Set("vhost", shovelInfo.Vhost)
	d.Set("info", []map[string]interface{}{info})

	return nil
}

func UpdateShovel(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	if d.HasChange("info") {
		_, newShovel := d.GetChange("info")

		newShovelList := newShovel.([]interface{})
		infoMap, ok := newShovelList[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("Unable to parse shovel info")
		}

		shovelDefinition := setShovelDefinition(infoMap).(rabbithole.ShovelDefinition)

		log.Printf("[DEBUG] RabbitMQ: Attempting to declare shovel %s in vhost %s", name, vhost)
		resp, err := rmqc.DeclareShovel(vhost, name, shovelDefinition)
		log.Printf("[DEBUG] RabbitMQ: shovel declartion response: %#v", resp)
		if err != nil {
			return err
		}
	}
	return ReadShovel(d, meta)
}

func DeleteShovel(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to delete shovel %s", d.Id())

	resp, err := rmqc.DeleteShovel(vhost, name)
	log.Printf("[DEBUG] RabbitMQ: shovel deletion response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error deleting RabbitMQ shovel: %s", resp.Status)
	}

	return nil
}

func setShovelDefinition(shovelMap map[string]interface{}) interface{} {
	shovelDefinition := &rabbithole.ShovelDefinition{}

	if v, ok := shovelMap["ack_mode"].(string); ok {
		shovelDefinition.AckMode = v
	}

	if v, ok := shovelMap["add_forward_headers"].(bool); ok {
		shovelDefinition.AddForwardHeaders = v
	}

	if v, ok := shovelMap["delete_after"].(string); ok {
		shovelDefinition.DeleteAfter = rabbithole.DeleteAfter(v)
	}

	if v, ok := shovelMap["destination_add_forward_headers"].(bool); ok {
		shovelDefinition.DestinationAddForwardHeaders = v
	}

	if v, ok := shovelMap["destination_add_timestamp_header"].(bool); ok {
		shovelDefinition.DestinationAddTimestampHeader = v
	}

	if v, ok := shovelMap["destination_address"].(string); ok {
		shovelDefinition.DestinationAddress = v
	}

	if v, ok := shovelMap["destination_application_properties"].(map[string]interface{}); ok {
		shovelDefinition.DestinationApplicationProperties = v
	}

	if v, ok := shovelMap["destination_exchange"].(string); ok {
		shovelDefinition.DestinationExchange = v
	}

	if v, ok := shovelMap["destination_exchange_key"].(string); ok {
		shovelDefinition.DestinationExchangeKey = v
	}

	if v, ok := shovelMap["destination_properties"].(map[string]interface{}); ok {
		shovelDefinition.DestinationProperties = v
	}

	if v, ok := shovelMap["destination_protocol"].(string); ok {
		shovelDefinition.DestinationProtocol = v
	}

	if v, ok := shovelMap["destination_publish_properties"].(map[string]interface{}); ok {
		shovelDefinition.DestinationPublishProperties = v
	}

	if v, ok := shovelMap["destination_queue"].(string); ok {
		shovelDefinition.DestinationQueue = v
	}

	if v, ok := shovelMap["destination_uri"].(string); ok {
		shovelDefinition.DestinationURI = []string{v}
	}

	if v, ok := shovelMap["prefetch_count"].(int); ok {
		shovelDefinition.PrefetchCount = v
	}

	if v, ok := shovelMap["reconnect_delay"].(int); ok {
		shovelDefinition.ReconnectDelay = v
	}
	if v, ok := shovelMap["source_address"].(string); ok {
		shovelDefinition.SourceAddress = v
	}

	if v, ok := shovelMap["source_delete_after"].(string); ok {
		shovelDefinition.SourceDeleteAfter = rabbithole.DeleteAfter(v)
	}

	if v, ok := shovelMap["source_exchange"].(string); ok {
		shovelDefinition.SourceExchange = v
	}

	if v, ok := shovelMap["source_exchange_key"].(string); ok {
		shovelDefinition.SourceExchangeKey = v
	}
	if v, ok := shovelMap["source_prefetch_count"].(int); ok {
		shovelDefinition.SourcePrefetchCount = v
	}

	if v, ok := shovelMap["source_protocol"].(string); ok {
		shovelDefinition.SourceProtocol = v
	}

	if v, ok := shovelMap["source_queue"].(string); ok {
		shovelDefinition.SourceQueue = v
	}

	if v, ok := shovelMap["source_uri"].(string); ok {
		shovelDefinition.SourceURI = []string{v}
	}

	return *shovelDefinition
}
