package rabbitmq

import (
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBinding_importBasic(t *testing.T) {
	resourceName := "rabbitmq_binding.test"
	var bindingInfo rabbithole.BindingInfo

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccBindingConfig_basic,
				Check: testAccBindingCheck(
					resourceName, &bindingInfo,
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
