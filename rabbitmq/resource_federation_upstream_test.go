package rabbitmq

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func TestAccFederationUpstream(t *testing.T) {
	var upstream rabbithole.FederationUpstream
	resourceName := "rabbitmq_federation_upstream.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFederationUpstreamCheckDestroy(&upstream),
		Steps: []resource.TestStep{
			{
				Config: testAccFederationUpstream_create(),
				Check: resource.ComposeTestCheckFunc(
					testAccFederationUpstreamCheck(resourceName, &upstream),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.uri", "amqp://server-name"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.prefetch_count", "1000"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.reconnect_delay", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.ack_mode", "on-confirm"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.trust_user_id", "false"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.max_hops", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.expires", "0"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.message_ttl", "0"),
				)},
			{
				Config: testAccFederationUpstream_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccFederationUpstreamCheck(resourceName, &upstream),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.prefetch_count", "500"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.reconnect_delay", "10"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.ack_mode", "on-publish"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.trust_user_id", "true"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.max_hops", "2"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.expires", "1800000"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.message_ttl", "60000"),
				)},
		},
	})
}

func TestAccFederationUpstream_hasComponent(t *testing.T) {
	var upstream rabbithole.FederationUpstream
	resourceName := "rabbitmq_federation_upstream.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFederationUpstreamCheckDestroy(&upstream),
		Steps: []resource.TestStep{
			{
				Config: testAccFederationUpstream_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccFederationUpstreamCheck(resourceName, &upstream),
					resource.TestCheckResourceAttr(resourceName, "component", "federation-upstream"),
				),
			},
		},
	})
}

func TestAccFederationUpstream_defaults(t *testing.T) {
	var upstream rabbithole.FederationUpstream
	resourceName := "rabbitmq_federation_upstream.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFederationUpstreamCheckDestroy(&upstream),
		Steps: []resource.TestStep{
			{
				Config: testAccFederationUpstream_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccFederationUpstreamCheck(resourceName, &upstream),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.prefetch_count", "1000"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.reconnect_delay", "5"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.ack_mode", "on-confirm"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.trust_user_id", "false"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.max_hops", "1"),
				),
			},
		},
	})
}

func TestAccFederationUpstream_validation(t *testing.T) {
	var upstream rabbithole.FederationUpstream

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFederationUpstreamCheckDestroy(&upstream),
		Steps: []resource.TestStep{
			{
				Config:      testAccFederationUpstream_validation(),
				ExpectError: regexp.MustCompile("expected definition.0.ack_mode to be one of"),
			},
		},
	})
}

func testAccFederationUpstreamCheck(rn string, upstream *rabbithole.FederationUpstream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("federation upstream id not set")
		}

		id := strings.Split(rs.Primary.ID, "@")
		name := id[0]
		vhost := id[1]

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		upstreams, err := rmqc.ListFederationUpstreamsIn(vhost)
		if err != nil {
			return fmt.Errorf("Error retrieving federation upstreams: %s", err)
		}

		for _, up := range upstreams {
			if up.Name == name && up.Vhost == vhost {
				upstream = &up
				return nil
			}
		}

		return fmt.Errorf("Unable to find federation upstream %s", rn)
	}
}

func testAccFederationUpstreamCheckDestroy(upstream *rabbithole.FederationUpstream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)

		upstreams, err := rmqc.ListFederationUpstreamsIn(upstream.Vhost)
		if err != nil {
			return fmt.Errorf("Error retrieving federation upstreams: %s", err)
		}

		for _, up := range upstreams {
			if up.Name == upstream.Name && up.Vhost == upstream.Vhost {
				return fmt.Errorf("Federation upstream %s@%s still exists", upstream.Name, upstream.Vhost)
			}
		}

		return nil
	}
}

func testAccFederationUpstream_baseConfig() string {
	return `
resource "rabbitmq_vhost" "test" {
		name = "test"
}

resource "rabbitmq_permissions" "guest" {
		user = "guest"
		vhost = rabbitmq_vhost.test.name
		permissions {
				configure = ".*"
				write = ".*"
				read = ".*"
		}
}
`
}

func testAccFederationUpstream_create() string {
	return testAccFederationUpstream_baseConfig() + `
resource "rabbitmq_federation_upstream" "foo" {
		name = "foo"
		vhost = rabbitmq_permissions.guest.vhost

		definition {
				uri = "amqp://server-name"
				prefetch_count = 1000
				reconnect_delay = 1
				ack_mode = "on-confirm"
				trust_user_id = false

				exchange = ""
				max_hops = 1
				expires = 0
				message_ttl = 0

				queue = ""
		}
}
`
}

func testAccFederationUpstream_update() string {
	return testAccFederationUpstream_baseConfig() + `
resource "rabbitmq_federation_upstream" "foo" {
		name = "foo"
		vhost = rabbitmq_permissions.guest.vhost

		definition {
				uri = "amqp://server-name"
				prefetch_count = 500
				reconnect_delay = 10
				ack_mode = "on-publish"
				trust_user_id = true

				exchange = ""
				max_hops = 2
				expires = 1800000
				message_ttl = 60000

				queue = ""
		}
}
`
}

func testAccFederationUpstream_basic() string {
	return testAccFederationUpstream_baseConfig() + `
resource "rabbitmq_federation_upstream" "foo" {
		name = "foo"
		vhost = rabbitmq_permissions.guest.vhost

		definition {
				uri = "amqp://server-name"
		}
}
`
}

func testAccFederationUpstream_validation() string {
	return testAccFederationUpstream_baseConfig() + `
resource "rabbitmq_federation_upstream" "foo" {
		name = "foo"
		vhost = rabbitmq_permissions.guest.vhost

		definition {
				uri = "amqp://server-name"
				ack_mode = "not-valid"
		}
}
`
}
