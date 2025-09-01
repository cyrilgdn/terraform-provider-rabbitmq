package rabbitmq

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBinding_basic(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccBindingConfig_basic,
				Check: testAccBindingCheck(
					"rabbitmq_binding.test", &bindingInfo,
				),
			},
			{
				Config: testAccBindingConfig_basic,
				Check: testAccBindingCheck(
					"rabbitmq_binding.foo_to_bar", &bindingInfo,
				),
			},
		},
	})
}

func TestAccBinding_jsonArguments(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	js := `{"x-match": "all", "foo": ["bar", "baz"]}`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccBindingConfig_jsonArguments(js),
				Check: resource.ComposeTestCheckFunc(
					testAccBindingCheck("rabbitmq_binding.test", &bindingInfo),
					testAccBindingCheckJsonArguments("rabbitmq_binding.test", &bindingInfo, js),
				),
			},
		},
	})
}

func TestAccBinding_slashEscaping(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccBindingConfig_slashesAreOkay,
				Check: testAccBindingCheck(
					"rabbitmq_binding.test", &bindingInfo,
				),
			},
		},
	})
}

func TestAccBinding_propertiesKey(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			{
				Config: testAccBindingConfig_propertiesKey,
				Check: testAccBindingCheck(
					"rabbitmq_binding.test", &bindingInfo,
				),
			},
		},
	})
}

func testAccBindingCheck(rn string, bindingInfo *rabbithole.BindingInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]

		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("binding id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		bindingParts := strings.Split(rs.Primary.ID, "/")

		bindings, err := rmqc.ListBindingsIn(percentDecodeSlashes(bindingParts[0]))
		if err != nil {
			return fmt.Errorf("error retrieving exchange: %s", err)
		}

		for _, binding := range bindings {
			if binding.Source == bindingParts[1] && binding.Destination == bindingParts[2] && binding.DestinationType == bindingParts[3] && binding.PropertiesKey == bindingParts[4] {
				*bindingInfo = binding
				return nil
			}
		}

		return fmt.Errorf("unable to find binding %s", rn)
	}
}

func testAccBindingCheckJsonArguments(rn string, bindingInfo *rabbithole.BindingInfo, js string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var configMap map[string]any
		if err := json.Unmarshal([]byte(js), &configMap); err != nil {
			return err
		}
		if !reflect.DeepEqual(configMap, bindingInfo.Arguments) {
			return fmt.Errorf("Passed arguments does not match binding arguments")
		}

		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}
		var configMap2 map[string]any
		if err := json.Unmarshal([]byte(rs.Primary.Attributes["arguments_json"]), &configMap2); err != nil {
			return err
		}
		if !reflect.DeepEqual(configMap2, bindingInfo.Arguments) {
			return fmt.Errorf("Arguments in state does not match binding arguments")
		}

		return nil
	}
}

func testAccBindingCheckDestroy(bindingInfo rabbithole.BindingInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)

		bindings, err := rmqc.ListBindingsIn(bindingInfo.Vhost)
		if err != nil {
			return fmt.Errorf("error retrieving exchange: %s", err)
		}

		for _, binding := range bindings {
			if binding.Source == bindingInfo.Source && binding.Destination == bindingInfo.Destination && binding.DestinationType == bindingInfo.DestinationType && binding.PropertiesKey == bindingInfo.PropertiesKey {
				return fmt.Errorf("Binding still exists")
			}
		}

		return nil
	}
}

const testAccBindingConfig_basic = `
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

resource "rabbitmq_exchange" "foo" {
    name = "foo"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_exchange" "bar" {
    name = "foo"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_exchange" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "foo_to_bar" {
    source = "${rabbitmq_exchange.foo.name}"
    vhost = "${rabbitmq_vhost.test.name}"
    destination = "${rabbitmq_exchange.bar.name}"
    destination_type = "exchange"
    routing_key = "#"
}

resource "rabbitmq_binding" "test" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_vhost.test.name}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "#"
}`

func testAccBindingConfig_jsonArguments(j string) string {
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

resource "rabbitmq_exchange" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "test" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_vhost.test.name}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "#"
    arguments_json = "${var.arguments}"
}`, j)
}

const testAccBindingConfig_slashesAreOkay = `
resource "rabbitmq_vhost" "test" {
    name = "/virtual//host///"
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
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "topic"
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "test" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "///routing//key/"
    arguments = {
      key1 = "value1"
      key2 = "value2"
      key3 = "value3"
    }
}
`

const testAccBindingConfig_propertiesKey = `
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
    name = "Test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "topic"
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_queue" "test" {
    name = "Test.Queue"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "test" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_vhost.test.name}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "ANYTHING.#"
    arguments = {
      key1 = "value1"
      key2 = "value2"
      key3 = "value3"
    }
}
`
