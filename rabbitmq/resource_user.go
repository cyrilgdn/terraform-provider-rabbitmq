package rabbitmq

import (
	"fmt"
	"log"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/go-cty/cty"
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
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"password", "password_wo"},
			},

			"password_wo": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				WriteOnly:    true,
				ExactlyOneOf: []string{"password", "password_wo"},
				RequiredWith: []string{"password_wo_version"},
			},

			// password_wo is never stored in state, so Terraform can never detect a
			// change to it. Changing password_wo_version is what produces a diff and
			// therefore what triggers a password rotation.
			"password_wo_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"password_wo"},
			},

			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func CreateUser(d *schema.ResourceData, meta any) error {
	rmqc := meta.(*rabbithole.Client)

	name := d.Get("name").(string)

	password, err := userPassword(d)
	if err != nil {
		return err
	}

	userSettings := rabbithole.UserSettings{
		Password: password,
		Tags:     userTagsToString(d),
	}

	log.Printf("[DEBUG] RabbitMQ: Attempting to create user %s", name)

	resp, err := rmqc.PutUser(name, userSettings)
	log.Printf("[DEBUG] RabbitMQ: user creation response: %#v", resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error creating RabbitMQ user: %s", resp.Status)
	}

	d.SetId(name)

	return ReadUser(d, meta)
}

func ReadUser(d *schema.ResourceData, meta any) error {
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

func UpdateUser(d *schema.ResourceData, meta any) error {
	rmqc := meta.(*rabbithole.Client)

	name := d.Id()
	tags := userTagsToString(d)

	// The password is sent on every update, including one that only changes tags.
	// PUT /api/users replaces the whole user, and UserSettings.Password is `omitempty`,
	// so an empty value here omits the field and RabbitMQ resets password_hash to "" --
	// wiping the password and locking the user out, while still returning success.
	password, err := userPassword(d)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("error updating RabbitMQ user: %s", resp.Status)
	}

	return ReadUser(d, meta)
}

func DeleteUser(d *schema.ResourceData, meta any) error {
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
		return fmt.Errorf("error deleting RabbitMQ user: %s", resp.Status)
	}

	return nil
}

// userPassword resolves the password to send to RabbitMQ. The schema's ExactlyOneOf
// guarantees that exactly one of password and password_wo is set in configuration.
//
// password_wo is a write-only attribute, so it is never persisted to state and
// d.Get("password_wo") always returns the zero value. It has to be read back out of
// the raw configuration instead.
func userPassword(d *schema.ResourceData) (string, error) {
	password := d.Get("password").(string)

	if !d.GetRawConfig().IsNull() {
		v, diags := d.GetRawConfigAt(cty.GetAttrPath("password_wo"))
		if diags.HasError() {
			return "", fmt.Errorf("error reading password_wo from configuration: %s: %s", diags[0].Summary, diags[0].Detail)
		}
		if v.IsKnown() && !v.IsNull() {
			password = v.AsString()
		}
	}

	// An empty password is allowed and meaningful: UserSettings.Password is
	// `omitempty`, so it is dropped from the request body and RabbitMQ stores the
	// user with no usable password hash. The internal backend then refuses every
	// password for that user, which is how a user is set up to authenticate through
	// a later backend in the auth_backends chain (LDAP, x509) instead.
	return password, nil
}

func userTagsToString(d *schema.ResourceData) rabbithole.UserTags {
	tagList := rabbithole.UserTags{}
	for _, v := range d.Get("tags").([]any) {
		if tag, ok := v.(string); ok {
			tagList = append(tagList, tag)
		}
	}

	return tagList
}
