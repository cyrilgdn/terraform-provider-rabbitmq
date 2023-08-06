package rabbitmq

import "testing"

func TestParseVHostResourceIdString(t *testing.T) {
	var badInputs = []string{
		"",
		"foo/test",
		"footest",
		"foo@bar@test",
	}

	for _, input := range badInputs {
		_, _, err := parseVHostResourceIdString(input)
		if err == nil {
			t.Errorf("parseId failed for: %s.", input)
		}
	}

	var goodInputs = []struct {
		input string
		name  string
		vhost string
	}{
		{"foo@test", "foo", "test"},
		{"foo@/", "foo", "/"},
		{"foo/bar/baz@/", "foo/bar/baz", "/"},
		{"foo%40bar@test", "foo@bar", "test"},
		{"foo@bar%40test", "foo", "bar@test"},
		{"foo%40bar@my%40test", "foo@bar", "my@test"},
	}

	for _, test := range goodInputs {
		name, vhost, err := parseVHostResourceIdString(test.input)
		if err != nil || name != test.name || vhost != test.vhost {
			t.Errorf("parseId failed for: %s.", test.input)
		}
	}
}
