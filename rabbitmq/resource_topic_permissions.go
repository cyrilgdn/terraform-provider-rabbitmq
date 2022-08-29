package rabbitmq

import (
	"fmt"
	"log"
	"strconv"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTopicPermissions() *schema.Resource {
	return &schema.Resource{
		Create: CreateTopicPermissions,
		Update: UpdateTopicPermissions,
		Read:   ReadTopicPermissions,
		Delete: DeleteTopicPermissions,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vhost": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},

			"permissions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exchange": {
							Type:     schema.TypeString,
							Required: true,
						},

						"write": {
							Type:     schema.TypeString,
							Required: true,
						},

						"read": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

// CreateTopicPermissions for given exchanges
func CreateTopicPermissions(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	user := d.Get("user").(string)
	vhost := d.Get("vhost").(string)
	permsSet := d.Get("permissions").(*schema.Set)

	for _, exchange := range permsSet.List() {

		permsMap, ok := exchange.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Unable to parse permissions")
		}

		if err := setTopicPermissionsIn(rmqc, vhost, user, permsMap); err != nil {
			return err
		}
	}

	id := fmt.Sprintf("%s@%s", user, vhost)
	d.SetId(id)

	return ReadTopicPermissions(d, meta)
}

// ReadTopicPermissions for the given ID
func ReadTopicPermissions(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	user, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	userPerms, err := rmqc.GetTopicPermissionsIn(vhost, user)
	if err != nil {
		return checkDeleted(d, err)
	}

	log.Printf("[DEBUG] RabbitMQ: Topic permission retrieved for %s: %#v", d.Id(), userPerms)

	d.Set("user", userPerms[0].User)
	d.Set("vhost", userPerms[0].Vhost)

	perms := make([]map[string]interface{}, len(userPerms))
	for i, perm := range userPerms {
		p := make(map[string]interface{})
		p["exchange"] = perm.Exchange
		p["write"] = perm.Write
		p["read"] = perm.Read
		perms[i] = p
	}
	d.Set("permissions", perms)

	return nil
}

// UpdateTopicPermissions for given ID
func UpdateTopicPermissions(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	user, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	if d.HasChange("permissions") {
		err = DeleteTopicPermissions(d, meta)
		if err != nil {
			return err
		}
		_, newPerms := d.GetChange("permissions")
		newPermsSet := newPerms.(*schema.Set)
		for _, exchange := range newPermsSet.List() {
			permsMap, ok := exchange.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Unable to parse permissions")
			}

			if err := setTopicPermissionsIn(rmqc, vhost, user, permsMap); err != nil {
				return err
			}
		}
	}

	return ReadTopicPermissions(d, meta)
}

// DeleteTopicPermissions for given ID
func DeleteTopicPermissions(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	user, vhost, err := parseResourceId(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to delete topic permission for %s", d.Id())

	resp, err := rmqc.ClearTopicPermissionsIn(vhost, user)
	log.Printf("[DEBUG] RabbitMQ: Topic permission delete response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		// The permissions were already deleted
		return nil
	}

	if resp.StatusCode >= 400 {
		verErr := checkVersion(rmqc)
		if verErr != nil {
			return verErr
		}
		return fmt.Errorf("Error deleting RabbitMQ topic permission: %s", resp.Status)
	}

	return nil
}

func setTopicPermissionsIn(rmqc *rabbithole.Client, vhost string, user string, permsMap map[string]interface{}) error {
	perms := rabbithole.TopicPermissions{}

	if v, ok := permsMap["exchange"].(string); ok {
		perms.Exchange = v
	}

	if v, ok := permsMap["write"].(string); ok {
		perms.Write = v
	}

	if v, ok := permsMap["read"].(string); ok {
		perms.Read = v
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to set topic permissions for %s@%s: %#v", user, vhost, perms)

	resp, err := rmqc.UpdateTopicPermissionsIn(vhost, user, perms)
	log.Printf("[DEBUG] RabbitMQ: Permission response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		verErr := checkVersion(rmqc)
		if verErr != nil {
			return verErr
		}
		return fmt.Errorf("Error setting topic permissions: %s", resp.Status)
	}

	return nil
}

func checkVersion(rmqc *rabbithole.Client) error {
	overview, _ := rmqc.Overview()
	ver, _ := strconv.ParseFloat(overview.RabbitMQVersion, 32)
	if ver < 3.7 {
		return fmt.Errorf("Topic permissions were adding in RabbitMQ 3.7, connected to %s", overview.RabbitMQVersion)
	}
	return nil
}
