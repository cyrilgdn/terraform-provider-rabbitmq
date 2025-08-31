package rabbitmq

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func checkDeleted(d *schema.ResourceData, err error) error {
	var errorResponse rabbithole.ErrorResponse
	if errors.As(err, &errorResponse) {
		if errorResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
	}
	return err
}

// Because slashes are used to separate different components when constructing binding IDs,
// we need a way to ensure any components that include slashes can survive the round trip.
// Percent-encoding is a straightforward way of doing so.
// (reference: https://developer.mozilla.org/en-US/docs/Glossary/percent-encoding)

func percentEncodeSlashes(s string) string {
	// Encode any percent signs, then encode any forward slashes.
	return strings.ReplaceAll(strings.ReplaceAll(s, "%", "%25"), "/", "%2F")
}

func percentDecodeSlashes(s string) string {
	// Decode any forward slashes, then decode any percent signs.
	return strings.ReplaceAll(strings.ReplaceAll(s, "%2F", "/"), "%25", "%")
}

// get the id of the resource from the ResourceData
func parseResourceId(d *schema.ResourceData) (name, vhost string, err error) {
	return parseId(d.Id())
}

// get the resource name and rabbitmq vhost from the resource id
func parseId(resourceId string) (name, vhost string, err error) {
	parts := strings.Split(resourceId, "@")
	if len(parts) != 2 {
		err = fmt.Errorf("unable to parse resource id: %s", resourceId)
		return
	}
	name = parts[0]
	vhost = parts[1]
	return
}
