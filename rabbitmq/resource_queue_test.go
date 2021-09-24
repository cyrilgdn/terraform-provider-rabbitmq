package rabbitmq

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func TestAccQueue_basic(t *testing.T) {
	var queueInfo rabbithole.QueueInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccQueueCheckDestroy(&queueInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic,
				Check: testAccQueueCheck(
					"rabbitmq_queue.test", &queueInfo,
				),
			},
			{
				Config: testAccQueueConfig_update,
				Check: testAccQueueCheck(
					"rabbitmq_queue.test", &queueInfo,
				),
			},
		},
	})
}

func TestAccQueue_jsonArguments(t *testing.T) {
	var queueInfo rabbithole.QueueInfo
	js := `{"x-message-ttl": 5000,"foo": "bar","baz": 50}`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccQueueCheckDestroy(&queueInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_jsonArguments(js),
				Check: resource.ComposeTestCheckFunc(
					testAccQueueCheck("rabbitmq_queue.test", &queueInfo),
					testAccQueueCheckJsonArguments("rabbitmq_queue.test", &queueInfo, js),
				),
			},
		},
	})
}

func testAccQueueCheck(rn string, queueInfo *rabbithole.QueueInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("queue id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		queueParts := strings.Split(rs.Primary.ID, "@")

		queues, err := rmqc.ListQueuesIn(queueParts[1])
		if err != nil {
			return fmt.Errorf("Error retrieving queue: %s", err)
		}

		for _, queue := range queues {
			if queue.Name == queueParts[0] && queue.Vhost == queueParts[1] {
				*queueInfo = queue
				return nil
			}
		}

		return fmt.Errorf("Unable to find queue %s", rn)
	}
}

func testAccQueueCheckJsonArguments(rn string, queueInfo *rabbithole.QueueInfo, js string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(js), &configMap); err != nil {
			return err
		}
		if !reflect.DeepEqual(configMap, queueInfo.Arguments) {
			return fmt.Errorf("Passed arguments does not match queue arguments")
		}

		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}
		var configMap2 map[string]interface{}
		if err := json.Unmarshal([]byte(rs.Primary.Attributes["settings.0.arguments_json"]), &configMap2); err != nil {
			return err
		}
		if !reflect.DeepEqual(configMap2, queueInfo.Arguments) {
			return fmt.Errorf("Arguments in state does not match queue arguments")
		}

		return nil
	}
}

func testAccQueueCheckDestroy(queueInfo *rabbithole.QueueInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)

		queues, err := rmqc.ListQueuesIn(queueInfo.Vhost)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return fmt.Errorf("Error retrieving queues: %s", err)
		}

		for _, queue := range queues {
			if queue.Name == queueInfo.Name && queue.Vhost == queueInfo.Vhost {
				return fmt.Errorf("Queue %s@%s still exist", queueInfo.Name, queueInfo.Vhost)
			}
		}

		return nil
	}
}

const testAccQueueConfig_basic = `
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

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = false
        auto_delete = true
    }
}`

const testAccQueueConfig_update = `
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

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}`

func testAccQueueConfig_jsonArguments(j string) string {
	return fmt.Sprintf(`
variable "arguments" {
	default = <<EOF
%s
EOF
}

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

resource "rabbitmq_queue" "test" {
	name = "test"
	vhost = "${rabbitmq_permissions.guest.vhost}"
	settings {
		durable = false
		auto_delete = true
		arguments_json = "${var.arguments}"
	}
}`, j)
}
