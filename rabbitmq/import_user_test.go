package rabbitmq

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUser_importBasic(t *testing.T) {
	resourceName := "rabbitmq_user.test"
	var user string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic,
				Check: testAccUserCheck(
					resourceName, &user,
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccUser_importPasswordWO(t *testing.T) {
	resourceName := "rabbitmq_user.test"
	var user string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordWO_v1,
				Check:  testAccUserCheck(resourceName, &user),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// password_wo is write-only so it is never in state and needs no ignore.
				// password_wo_version is real state, but RabbitMQ has no notion of it, so
				// an imported user cannot report one. The first plan after an import will
				// therefore set it, which triggers an update that asserts the configured
				// password -- the intended convergence.
				ImportStateVerifyIgnore: []string{"password_wo_version"},
			},
		},
	})
}
