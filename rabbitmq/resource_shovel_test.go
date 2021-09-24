package rabbitmq

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func TestAccShovel(t *testing.T) {
	var shovelInfo rabbithole.ShovelInfo

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccShovelCheckDestroy(&shovelInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccShovelConfig_basic,
				Check: testAccShovelCheck(
					"rabbitmq_shovel.shovelTest", &shovelInfo,
				),
			},
		},
	})
}

func testAccShovelCheck(rn string, shovelInfo *rabbithole.ShovelInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("shovel id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		shovelParts := strings.Split(rs.Primary.ID, "@")

		shovelInfos, err := rmqc.ListShovels()
		if err != nil {
			return fmt.Errorf("Error retrieving shovels: %s", err)
		}

		for _, info := range shovelInfos {
			if info.Name == shovelParts[0] && info.Vhost == shovelParts[1] {
				shovelInfo = &info
				return nil
			}
		}

		return fmt.Errorf("Unable to find shovel %s", rn)
	}
}

func testAccShovelCheckDestroy(shovelInfo *rabbithole.ShovelInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)

		shovelInfos, err := rmqc.ListShovels()
		if err != nil {
			return fmt.Errorf("Error retrieving shovels: %s", err)
		}

		for _, info := range shovelInfos {
			if info.Name == shovelInfo.Name && info.Vhost == shovelInfo.Vhost {
				return fmt.Errorf("shovel still exists: %v", info)
			}
		}

		return nil
	}
}

const testAccShovelConfig_basic = `
resource "rabbitmq_vhost" "test" {
    name = "test"
}

resource "rabbitmq_permissions" "guest" {
    user = "guest"
    vhost = "${rabbitmq_vhost.test.name}"
    permissions {
        configure = ".*"
        write = ".*"
        read = ".*"
    }
}

resource "rabbitmq_exchange" "test" {
    name = "test_exchange"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_queue" "test" {
	name = "test_queue"
	vhost = "${rabbitmq_exchange.test.vhost}"
	settings {
		durable = false
		auto_delete = true
	}
}

resource "rabbitmq_shovel" "shovelTest" {
	name = "shovelTest"
	vhost = "${rabbitmq_queue.test.vhost}"
	info {
		source_uri = "amqp:///test"
		source_exchange = "${rabbitmq_exchange.test.name}"
		source_exchange_key = "test"
		destination_uri = "amqp:///test"
		destination_queue = "${rabbitmq_queue.test.name}"
	}
}`
