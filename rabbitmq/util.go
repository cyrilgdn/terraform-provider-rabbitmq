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
	return strings.Replace(strings.Replace(s, "%", "%25", -1), "/", "%2F", -1)
}

func percentDecodeSlashes(s string) string {
	// Decode any forward slashes, then decode any percent signs.
	return strings.Replace(strings.Replace(s, "%2F", "/", -1), "%25", "%", -1)
}

// Builds a combined resource id using a percent encoded name and vhost.
func buildVHostResourceId(name, vhost string) string {
	id := fmt.Sprintf("%s@%s", percentEncodeAtSymbols(name), percentEncodeAtSymbols(vhost))
	return id
}

// Get the resource name and rabbitmq vhost from the ResourceData.
func parseVHostResourceId(d *schema.ResourceData) (name, vhost string, err error) {
	return parseVHostResourceIdString(d.Id())
}

// Get the resource name and rabbitmq vhost from the resource id.
func parseVHostResourceIdString(resourceId string) (name, vhost string, err error) {
	parts := strings.Split(resourceId, "@")
	if len(parts) != 2 {
		err = fmt.Errorf("Unable to parse resource id: %s", resourceId)
		return
	}
	name = percentDecodeAtSymbols(parts[0])
	vhost = percentDecodeAtSymbols(parts[1])
	return
}

// Because the @ symbol is used to separate the name & vhost components when building a "vhost resource id",
// we need a way to ensure that any @ symbol within the components can survive the round trip.
// Percent-encoding is a straightforward way of doing so.
// (reference: https://developer.mozilla.org/en-US/docs/Glossary/percent-encoding)

func percentEncodeAtSymbols(s string) string {
	// Encode any percent signs, then encode any @ symbols.
	return strings.Replace(strings.Replace(s, "%", "%25", -1), "@", "%40", -1)
}

func percentDecodeAtSymbols(s string) string {
	// Decode any @ symbols, then decode any percent signs.
	return strings.Replace(strings.Replace(s, "%40", "@", -1), "%25", "%", -1)
}
