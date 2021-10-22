package rabbitmq

import (
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFederationUpstream_importBasic(t *testing.T) {
	var upstream rabbithole.FederationUpstream
	resourceName := "rabbitmq_federation_upstream.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFederationUpstreamCheckDestroy(&upstream),
		Steps: []resource.TestStep{
			{
				Config: testAccFederationUpstream_create(),
				Check: testAccFederationUpstreamCheck(
					resourceName, &upstream,
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
