package rabbitmq

import (
	"os"
	"regexp"
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTopicPermissions_importBasic(t *testing.T) {
	resourceName := "rabbitmq_topic_permissions.test"
	var topicPermissionInfo rabbithole.TopicPermissionInfo
	var expectErr0 *regexp.Regexp
	var expectErr1 *regexp.Regexp
	checkDestroy := testAccTopicPermissionsCheckDestroy(&topicPermissionInfo)
	if os.Getenv("RABBITMQ_VERSION") == "3.6" {
		expectErr0, _ = regexp.Compile("^errors during apply: Topic permissions were adding in RabbitMQ 3.7, connected to 3.6.*$")
		expectErr1, _ = regexp.Compile("^Resource specified by ResourceName couldn't be found: rabbitmq_topic_permissions.test$")
		checkDestroy = nil
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPermissionsConfigBasic,
				Check: testAccTopicPermissionsCheck(
					resourceName, &topicPermissionInfo,
				),
				ExpectError: expectErr0,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       expectErr1,
			},
		},
	})
}
