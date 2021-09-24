package rabbitmq

import (
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccExchange_importBasic(t *testing.T) {
	resourceName := "rabbitmq_exchange.test"
	var exchangeInfo rabbithole.ExchangeInfo

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccExchangeCheckDestroy(&exchangeInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccExchangeConfig_basic,
				Check: testAccExchangeCheck(
					resourceName, &exchangeInfo,
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
