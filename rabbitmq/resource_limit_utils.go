package rabbitmq

import (
	"fmt"
	"strings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

func mergeLimits(source map[string]int, destination map[string]int) {

	for k, v := range source {

		destination[k] = v
	}
}

/*
Iterates over an array with
structs representing user or host limits
and forms a single dictionary from them.
*/
func getLimits(remoteLimits interface{}, err error) (map[string]int, error) {

	if err != nil {

		return nil, err
	}

	limits := make(map[string]int)

	switch data := remoteLimits.(type) {

	case []rabbithole.UserLimitsInfo:
		for _, v := range data {

			mergeLimits(v.Value, limits)
		}
	case []rabbithole.VhostLimitsInfo:
		for _, v := range data {

			mergeLimits(v.Value, limits)
		}
	}

	return limits, err
}

/*
Parses the rabbitmq_limit identifier of a resource represented
in the format {scope}@{limit}@{alias} and returns each segment
as separate variables.
*/
func parseLimitID(resourceId string) (string, string, string, error) {

	segments := strings.Split(resourceId, "@")

	if len(segments) < 3 {

		return "", "", "", fmt.Errorf("unable to determine limit ID for resource %s", resourceId)
	}

	scope := segments[0]
	limit := segments[1]
	alias := segments[2]

	return scope, limit, alias, nil
}
