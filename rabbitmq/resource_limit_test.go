package rabbitmq

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

const testLimitUser = `
resource "rabbitmq_limit" "limit_user" {

	scope = "user"
	alias = "guest"
	limit = "max-channels"
	value = 200
 }`

const testLimitUser_update = `
resource "rabbitmq_limit" "limit_user" {

	scope = "user"
	alias = "guest"
	limit = "max-channels"
	value = 200
}`

const testLimitVhost_basic = `
resource "rabbitmq_limit" "limit_vhost" {

	scope = "vhost"
	alias = "/"
	limit = "max-connections"
	value = 100
 }`

const testLimitVhost_update = `
resource "rabbitmq_limit" "limit_vhost" {

	scope = "vhost"
	alias = "/"
	limit = "max-connections"
	value = 200
 }`

/*
Signature of the predicate that checks
the dictionary for the presence of the
given key.
*/
type ValidateDictionary func(limit string, limits map[string]int) error

func TestLimitUser_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{

		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testLimitDestroy("rabbitmq_limit.limit_user"),

		Steps: []resource.TestStep{
			{
				Config: testLimitUser,
				Check:  testLimitCheck("rabbitmq_limit.limit_user"),
			},
			{
				Config: testLimitUser_update,
				Check:  testLimitCheck("rabbitmq_limit.limit_user"),
			},
		},
	})
}

func TestLimitVhost_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{

		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testLimitDestroy("rabbitmq_limit.limit_vhost"),

		Steps: []resource.TestStep{
			{
				Config: testLimitVhost_basic,
				Check:  testLimitCheck("rabbitmq_limit.limit_vhost"),
			},
			{
				Config: testLimitVhost_update,
				Check:  testLimitCheck("rabbitmq_limit.limit_vhost"),
			},
		},
	})
}

/*
Returns the component parts from which the
identifier of the resource-limit is formed.
*/
func getLimitParts(state *terraform.State, resource string) (string, string, string, error) {

	rs, ok := state.RootModule().Resources[resource]

	if !ok || rs.Primary.ID == "" {

		return "", "", "", fmt.Errorf("resource %s not found or id not set", resource)
	}

	return parseLimitID(rs.Primary.ID)
}

func testLimit(s *terraform.State, resource string, predicat ValidateDictionary) error {

	scope, limit, alias, err := getLimitParts(s, resource)

	if err != nil {

		return err
	}

	rmqc := testAccProvider.Meta().(*rabbithole.Client)

	var limits map[string]int

	switch scope {

	case "user":
		limits, err = getLimits(rmqc.GetUserLimits(alias))
	case "vhost":
		limits, err = getLimits(rmqc.GetVhostLimits(alias))
	}

	if err == nil {

		return predicat(limit, limits)
	}

	return err
}

func testLimitCheck(resource string) resource.TestCheckFunc {

	return func(state *terraform.State) error {

		return testLimit(state, resource, func(limit string, limits map[string]int) error {

			if _, ok := limits[limit]; ok {

				return nil
			}

			return fmt.Errorf("unable to find limit %s", limit)
		})
	}
}

func testLimitDestroy(resource string) resource.TestCheckFunc {

	return func(state *terraform.State) error {

		return testLimit(state, resource, func(limit string, limits map[string]int) error {

			if _, ok := limits[limit]; ok {

				return fmt.Errorf("limit %s still exists", limit)
			}

			return nil
		})
	}
}
