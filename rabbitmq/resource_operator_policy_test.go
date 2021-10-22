package rabbitmq

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func TestAccOperatorPolicy(t *testing.T) {
	var operatorPolicy rabbithole.OperatorPolicy
	resourceName := "rabbitmq_operator_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccOperatorPolicyCheckDestroy(&operatorPolicy),
		Steps: []resource.TestStep{
			{
				Config: testAccOperatorPolicyConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccOperatorPolicyCheck(resourceName, &operatorPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.pattern", ".*"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.apply_to", "queues"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.definition.max-length", "10000"),
				),
			},
			{
				Config: testAccOperatorPolicyConfig_update,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccOperatorPolicyCheck(resourceName, &operatorPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.pattern", ".*"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.apply_to", "queues"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.definition.max-length", "202"),
					resource.TestCheckResourceAttr(resourceName, "policy.0.definition.expires", "60000"),
				),
			},
		},
	})
}

func testAccOperatorPolicyCheck(rn string, operatorPolicy *rabbithole.OperatorPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("operator policy id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		operatorPolicyParts := strings.Split(rs.Primary.ID, "@")

		operatorPolicies, err := rmqc.ListOperatorPolicies()
		if err != nil {
			return fmt.Errorf("Error retrieving operator policies: %s", err)
		}

		for _, p := range operatorPolicies {
			if p.Name == operatorPolicyParts[0] && p.Vhost == operatorPolicyParts[1] {
				operatorPolicy = &p
				return nil
			}
		}

		return fmt.Errorf("Unable to find operator policy %s", rn)
	}
}

func testAccOperatorPolicyCheckDestroy(operatorPolicy *rabbithole.OperatorPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)

		operatorPolicies, err := rmqc.ListOperatorPolicies()
		if err != nil {
			return fmt.Errorf("Error retrieving operator policies: %s", err)
		}

		for _, p := range operatorPolicies {
			if p.Name == operatorPolicy.Name && p.Vhost == operatorPolicy.Vhost {
				return fmt.Errorf("OperatorPolicy %s@%s still exist", operatorPolicy.Name, operatorPolicy.Vhost)
			}
		}

		return nil
	}
}

const testAccOperatorPolicyConfig_basic = `
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

resource "rabbitmq_operator_policy" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    policy {
        pattern = ".*"
        priority = 1
        apply_to = "queues"
        definition = {
            max-length = 10000
        }
    }
}`

const testAccOperatorPolicyConfig_update = `
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

resource "rabbitmq_operator_policy" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    policy {
        pattern = ".*"
        priority = 1
        apply_to = "queues"
        definition = {
            max-length = 202
			expires = 60000
        }
    }
}`
