package rabbitmq

import (
	"fmt"
	"log"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: CreateUser,
		Update: UpdateUser,
		Read:   ReadUser,
		Delete: DeleteUser,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func CreateUser(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name := d.Get("name").(string)

	userSettings := rabbithole.UserSettings{
		Password: d.Get("password").(string),
		Tags:     userTagsToString(d),
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to create user %s", name)

	// Check if already exists
	_, not_found := rmqc.GetUser(name)
	if not_found == nil {
		return fmt.Errorf("Error creating RabbitMQ user '%s': user already exists", name)
	}

	resp, err := rmqc.PutUser(name, userSettings)
	log.Printf("[DEBUG] RabbitMQ: user creation response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error creating RabbitMQ user: %s", resp.Status)
	}

	d.SetId(name)

	return ReadUser(d, meta)
}

func ReadUser(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	user, err := rmqc.GetUser(d.Id())
	if err != nil {
		return checkDeleted(d, err)
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

	return nil
}

func UpdateUser(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name := d.Id()
	tags := userTagsToString(d)
	password := d.Get("password").(string)

	userSettings := rabbithole.UserSettings{
		Password: password,
		Tags:     tags,
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to update user %s", name)

	resp, err := rmqc.PutUser(name, userSettings)
	log.Printf("[DEBUG] RabbitMQ: User update response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error updating RabbitMQ user: %s", resp.Status)
	}

	return ReadUser(d, meta)
}

func DeleteUser(d *schema.ResourceData, meta interface{}) error {
	rmqc := meta.(*rabbithole.Client)

	name := d.Id()
	log.Printf("[DEBUG] RabbitMQ: Attempting to delete user %s", name)

	resp, err := rmqc.DeleteUser(name)
	log.Printf("[DEBUG] RabbitMQ: User delete response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		// the user was automatically deleted
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error deleting RabbitMQ user '%s': %s", name, resp.Status)
	}

	return nil
}

func userTagsToString(d *schema.ResourceData) rabbithole.UserTags {
	tagList := rabbithole.UserTags{}
	for _, v := range d.Get("tags").([]interface{}) {
		if tag, ok := v.(string); ok {
			tagList = append(tagList, tag)
		}
	}

	return tagList
}
