package rabbitmq

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOperatorPolicy() *schema.Resource {
	return &schema.Resource{
		Create: CreateOperatorPolicy,
		Update: UpdateOperatorPolicy,
		Read:   ReadOperatorPolicy,
		Delete: DeleteOperatorPolicy,
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

			"policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pattern": {
							Type:     schema.TypeString,
							Required: true,
						},

						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"apply_to": {
							Type:     schema.TypeString,
							Required: true,
						},

						"definition": {
							Type:     schema.TypeMap,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func CreateOperatorPolicy(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name := d.Get("name").(string)
	vhost := d.Get("vhost").(string)
	operatorPolicyList := d.Get("policy").([]interface{})

	operatorPolicyMap, ok := operatorPolicyList[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Unable to parse operator policy")
	}

	if err := putOperatorPolicy(rmqc, vhost, name, operatorPolicyMap); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s@%s", name, vhost))

	return ReadOperatorPolicy(d, meta)
}

func ReadOperatorPolicy(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	operatorPolicy, err := rmqc.GetOperatorPolicy(vhost, name)
	if err != nil {
		return checkDeleted(d, err)
	}

	log.Printf("[DEBUG] RabbitMQ: OperatorPolicy retrieved for %s: %#v", d.Id(), operatorPolicy)

	d.Set("name", operatorPolicy.Name)
	d.Set("vhost", operatorPolicy.Vhost)

	setOperatorPolicy := make([]map[string]interface{}, 1)
	p := make(map[string]interface{})
	p["pattern"] = operatorPolicy.Pattern
	p["priority"] = operatorPolicy.Priority
	p["apply_to"] = operatorPolicy.ApplyTo

	operatorPolicyDefinition := make(map[string]interface{})
	for key, value := range operatorPolicy.Definition {
		switch v := value.(type) {
		case float64:
			value = strconv.FormatFloat(v, 'f', -1, 64)
		case []interface{}:
			var nodes []string
			for _, node := range v {
				if n, ok := node.(string); ok {
					nodes = append(nodes, n)
				}
			}
			value = strings.Join(nodes, ",")
		}
		operatorPolicyDefinition[key] = value
	}
	p["definition"] = operatorPolicyDefinition
	setOperatorPolicy[0] = p

	d.Set("policy", setOperatorPolicy)

	return nil
}

func UpdateOperatorPolicy(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	if d.HasChange("policy") {
		_, newOperatorPolicy := d.GetChange("policy")

		operatorPolicyList := newOperatorPolicy.([]interface{})
		operatorPolicyMap, ok := operatorPolicyList[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("Unable to parse operator policy")
		}

		if err := putOperatorPolicy(rmqc, vhost, name, operatorPolicyMap); err != nil {
			return err
		}
	}

	return ReadOperatorPolicy(d, meta)
}

func DeleteOperatorPolicy(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to delete operator policy for %s", d.Id())

	resp, err := rmqc.DeleteOperatorPolicy(vhost, name)
	log.Printf("[DEBUG] RabbitMQ: OperatorPolicy delete response: %#v", resp)
	if err != nil {
		return fmt.Errorf("could not delete operator policy: %w", err)
	}

	if resp.StatusCode == 404 {
		// the operator policy was automatically deleted
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error deleting RabbitMQ operator policy: %s", resp.Status)
	}

	return nil
}

func putOperatorPolicy(rmqc *rabbithole.Client, vhost string, name string, operatorPolicyMap map[string]interface{}) error {
	operatorPolicy := rabbithole.OperatorPolicy{}
	operatorPolicy.Vhost = vhost
	operatorPolicy.Name = name

	if v, ok := operatorPolicyMap["pattern"].(string); ok {
		operatorPolicy.Pattern = v
	}

	if v, ok := operatorPolicyMap["priority"].(int); ok {
		operatorPolicy.Priority = v
	}

	if v, ok := operatorPolicyMap["apply_to"].(string); ok {
		operatorPolicy.ApplyTo = v
	}

	if v, ok := operatorPolicyMap["definition"].(map[string]interface{}); ok {
		// special case for integers
		for key, val := range v {
			if x, ok := val.(string); ok {
				if x, err := strconv.ParseInt(x, 10, 64); err == nil {
					v[key] = x
				}
			}
		}

		operatorPolicy.Definition = v
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to declare operator policy for %s@%s: %#v", name, vhost, operatorPolicy)

	resp, err := rmqc.PutOperatorPolicy(vhost, name, operatorPolicy)
	log.Printf("[DEBUG] RabbitMQ: OperatorPolicy declare response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error declaring RabbitMQ operator policy: %s", resp.Status)
	}

	return nil
}
