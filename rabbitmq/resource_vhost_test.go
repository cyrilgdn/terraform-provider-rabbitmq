package rabbitmq

import (
	"fmt"
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVhost(t *testing.T) {
	var vhost string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccVhostCheckDestroy(vhost),
		Steps: []resource.TestStep{
			{
				Config: testAccVhostConfig_basic,
				Check: testAccVhostCheck(
					"rabbitmq_vhost.test", &vhost,
				),
			},
			{
				// Test that, once a vhost has been created and stored in the
				// state, even if it disappears from the RabbitMQ cluster, it
				// would be created without error.
				PreConfig: forceDropVhost(&vhost),
				Config:    testAccVhostConfig_basic,
				Check: testAccVhostCheck(
					"rabbitmq_vhost.test", &vhost,
				),
			},
		},
	})
}

func forceDropVhost(vhost *string) func() {
	return func() {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		resp, err := rmqc.DeleteVhost(*vhost)
		if err != nil {
			fmt.Printf("unable to delete vhost: %v", err)
			return
		}

		// Should get 204 when the vhost has been deleted
		if resp.StatusCode != 204 {
			panic(fmt.Errorf("unable to delete vhost: %v", resp))
		}
	}
}

func testAccVhostCheck(rn string, name *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vhost id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		vhosts, err := rmqc.ListVhosts()
		if err != nil {
			return fmt.Errorf("Error retrieving vhosts: %s", err)
		}

		for _, vhost := range vhosts {
			if vhost.Name == rs.Primary.ID {
				*name = rs.Primary.ID
				return nil
			}
		}

		return fmt.Errorf("Unable to find vhost %s", rn)
	}
}

func testAccVhostCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		vhosts, err := rmqc.ListVhosts()
		if err != nil {
			return fmt.Errorf("Error retrieving vhosts: %s", err)
		}

		for _, vhost := range vhosts {
			if vhost.Name == name {
				return fmt.Errorf("vhost still exists: %v", vhost)
			}
		}

		return nil
	}
}

const testAccVhostConfig_basic = `
resource "rabbitmq_vhost" "test" {
    name = "test"
	description = "test description"
	tracing = true
	max_connections = 100
	max_queues = 200
}`
