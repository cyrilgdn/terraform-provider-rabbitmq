package rabbitmq

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// --- unit tests (no broker required); acceptance tests follow below ---

func TestResourceUserPasswordSchema(t *testing.T) {
	s := resourceUser().Schema

	for _, k := range []string{"password", "password_wo", "password_wo_version"} {
		if s[k] == nil {
			t.Fatalf("schema is missing %s", k)
		}
	}

	// Without WriteOnly the plaintext password is persisted to state, which defeats
	// the entire purpose of the attribute.
	if !s["password_wo"].WriteOnly {
		t.Error("password_wo must have WriteOnly set")
	}

	if !s["password_wo"].Sensitive {
		t.Error("password_wo must have Sensitive set")
	}

	// password can no longer be Required, because ExactlyOneOf is invalid alongside
	// Required. ExactlyOneOf preserves the guarantee that Required:true used to give.
	if s["password"].Required {
		t.Error("password must be Optional, not Required")
	}

	if !s["password"].Optional {
		t.Error("password must be Optional")
	}

	if !s["password"].Sensitive {
		t.Error("password must have Sensitive set")
	}

	wantExactlyOneOf := []string{"password", "password_wo"}
	for _, k := range wantExactlyOneOf {
		if !reflect.DeepEqual(s[k].ExactlyOneOf, wantExactlyOneOf) {
			t.Errorf("%s: expected ExactlyOneOf %v, got %v", k, wantExactlyOneOf, s[k].ExactlyOneOf)
		}
	}

	// password_wo and password_wo_version are all-or-nothing.
	if !reflect.DeepEqual(s["password_wo"].RequiredWith, []string{"password_wo_version"}) {
		t.Errorf("password_wo: expected RequiredWith [password_wo_version], got %v", s["password_wo"].RequiredWith)
	}

	if !reflect.DeepEqual(s["password_wo_version"].RequiredWith, []string{"password_wo"}) {
		t.Errorf("password_wo_version: expected RequiredWith [password_wo], got %v", s["password_wo_version"].RequiredWith)
	}

	if s["password_wo_version"].Type != schema.TypeInt {
		t.Errorf("password_wo_version: expected TypeInt, got %s", s["password_wo_version"].Type)
	}
}

func TestAccUser_basic(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic,
				Check: testAccUserCheck(
					"rabbitmq_user.test", &user,
				),
			},
			{
				Config: testAccUserConfig_update,
				Check: testAccUserCheck(
					"rabbitmq_user.test", &user,
				),
			},
		},
	})
}

func TestUpdateTags_password(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testUpdateTagsCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck(
						"rabbitmq_user.test", &user,
					),
					testAccUserConnect("mctest", "foobar"),
				),
			},
			{
				Config: testUpdateTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck(
						"rabbitmq_user.test", &user,
					),
					testAccUserConnect("mctest", "foobar"),
				),
			},
		},
	})
}

func TestAccUser_emptyTag(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_emptyTag_1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 0),
				),
			},
			{
				Config: testAccUserConfig_emptyTag_2,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 1),
				),
			},
			{
				Config: testAccUserConfig_emptyTag_1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 0),
				),
			},
		},
	})
}

func TestAccUser_noTags(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_noTags_1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 0),
				),
			},
			{
				Config: testAccUserConfig_noTags_2,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 1),
				),
			},
		},
	})
}

func TestAccUser_passwordChange(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordChange_1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 2),
				),
			},
			{
				Config: testAccUserConfig_passwordChange_2,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserCheckTagCount(&user, 2),
				),
			},
		},
	})
}

func TestAccUser_passwordWO(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordWO_v1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserConnect("mctest", "wo-secret-one"),
					// The entire point of the feature: the secret must not reach state.
					resource.TestCheckNoResourceAttr("rabbitmq_user.test", "password_wo"),
					resource.TestCheckResourceAttr("rabbitmq_user.test", "password_wo_version", "1"),
				),
			},
		},
	})
}

func TestAccUser_passwordWO_rotate(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordWO_v1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserConnect("mctest", "wo-secret-one"),
				),
			},
			{
				// Bumping password_wo_version is what makes Terraform plan an update at
				// all; UpdateUser then sends whatever password_wo currently holds.
				Config: testAccUserConfig_passwordWO_v2,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserConnect("mctest", "wo-secret-two"),
					testAccUserCannotConnect("mctest", "wo-secret-one"),
					resource.TestCheckResourceAttr("rabbitmq_user.test", "password_wo_version", "2"),
				),
			},
		},
	})
}

// An empty password is a supported configuration, not a mistake. RabbitMQ stores the
// user with no usable password hash, so the internal backend refuses every password and
// authentication falls through to the next backend in the auth_backends chain. That is
// how users who authenticate only via LDAP or x509 are declared.
func TestAccUser_emptyPassword(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_emptyPassword,
				Check: resource.ComposeTestCheckFunc(
					// The user must exist...
					testAccUserCheck("rabbitmq_user.test", &user),
					// ...but must not be able to authenticate against the internal
					// backend, with an empty password or any other.
					testAccUserCannotConnect("mctest", ""),
					testAccUserCannotConnect("mctest", "anything"),
				),
			},
		},
	})
}

func testAccUserCheck(rn string, name *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		users, err := rmqc.ListUsers()
		if err != nil {
			return fmt.Errorf("error retrieving users: %s", err)
		}

		for _, user := range users {
			if user.Name == rs.Primary.ID {
				*name = rs.Primary.ID
				return nil
			}
		}

		return fmt.Errorf("unable to find user %s", rn)
	}
}

func testAccUserConnect(username, password string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := rabbithole.NewClient("http://localhost:15672", username, password)
		if err != nil {
			return fmt.Errorf("could not create rmq client: %v", err)
		}

		_, err = client.Whoami()
		if err != nil {
			return fmt.Errorf("could not call whoami with username %s: %v", username, err)
		}
		return nil
	}
}

// testAccUserCannotConnect asserts that the given credentials are rejected.
//
// It treats any Whoami() error as an authentication failure and so cannot distinguish
// "wrong password" from "broker unreachable". Pair it with a preceding successful
// testAccUserConnect check in the same step, so an unreachable broker fails there first.
func testAccUserCannotConnect(username, password string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := rabbithole.NewClient("http://localhost:15672", username, password)
		if err != nil {
			return fmt.Errorf("could not create rmq client: %v", err)
		}

		if _, err = client.Whoami(); err == nil {
			return fmt.Errorf("expected authentication to fail for user %s, but it succeeded", username)
		}
		return nil
	}
}

func testAccUserCheckTagCount(name *string, tagCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		user, err := rmqc.GetUser(*name)
		if err != nil {
			return fmt.Errorf("error retrieving user: %s", err)
		}

		var tagList []string
		for _, v := range user.Tags {
			if v != "" {
				tagList = append(tagList, v)
			}
		}

		if len(tagList) != tagCount {
			return fmt.Errorf("Expected %d tags, user has %d", tagCount, len(tagList))
		}

		return nil
	}
}

func testAccUserCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		users, err := rmqc.ListUsers()
		if err != nil {
			return fmt.Errorf("error retrieving users: %s", err)
		}

		for _, user := range users {
			if user.Name == name {
				return fmt.Errorf("user still exists: %s", name)
			}
		}

		return nil
	}
}

const testAccUserConfig_basic = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = ["administrator", "management"]
}`

const testAccUserConfig_update = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobarry"
    tags = ["management"]
}`

const testUpdateTagsCreate = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = ["management"]
}`

const testUpdateTagsUpdate = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = ["monitoring"]
}`

const testAccUserConfig_emptyTag_1 = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = [""]
}`

const testAccUserConfig_emptyTag_2 = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = ["administrator"]
}`

const testAccUserConfig_noTags_1 = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
}`

const testAccUserConfig_noTags_2 = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = ["administrator"]
}`

const testAccUserConfig_passwordChange_1 = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobar"
    tags = ["administrator", "management"]
}`

const testAccUserConfig_passwordChange_2 = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    password = "foobarry"
    tags = ["administrator", "management"]
}`

const testAccUserConfig_passwordWO_v1 = `
resource "rabbitmq_user" "test" {
    name                = "mctest"
    password_wo         = "wo-secret-one"
    password_wo_version = 1
    tags                = ["management"]
}`

const testAccUserConfig_passwordWO_v2 = `
resource "rabbitmq_user" "test" {
    name                = "mctest"
    password_wo         = "wo-secret-two"
    password_wo_version = 2
    tags                = ["management"]
}`

const testAccUserConfig_passwordToWO_before = `
resource "rabbitmq_user" "test" {
    name     = "mctest"
    password = "foobar"
    tags     = ["management"]
}`

const testAccUserConfig_emptyPassword = `
resource "rabbitmq_user" "test" {
    name     = "mctest"
    password = ""
    tags     = ["management"]
}`

func TestAccUser_passwordWO_versionUnchanged(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordWO_v1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserConnect("mctest", "wo-secret-one"),
				),
			},
			{
				// password_wo changes but password_wo_version does not. Because write-only
				// values are absent from state, Terraform computes no diff and never calls
				// Update, so the password is deliberately NOT rotated. This is the
				// documented contract, not a bug -- the docs tell users to bump the version.
				//
				// Note this step pins Terraform's behaviour rather than the provider's:
				// UpdateUser is never reached, so no change to it can make this step fail.
				// Its value is documenting the contract, and catching a future CustomizeDiff
				// that started forcing updates on password_wo alone.
				Config: testAccUserConfig_passwordWO_v1_changedSecret,
				Check: resource.ComposeTestCheckFunc(
					testAccUserConnect("mctest", "wo-secret-one"),
					testAccUserCannotConnect("mctest", "wo-secret-ignored"),
				),
			},
		},
	})
}

func TestAccUser_passwordWO_tagsOnlyUpdate(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordWO_v1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserConnect("mctest", "wo-secret-one"),
				),
			},
			{
				// Only tags change, so password_wo_version is unchanged. UpdateUser must
				// STILL send the password: PUT /api/users replaces the whole user, and
				// omitting the password resets password_hash to "" -- RabbitMQ reports
				// success and the account can no longer authenticate.
				//
				// This is the ONLY test that catches that regression. Gating the password
				// on d.HasChange("password_wo_version") leaves TestAccUser_passwordWO_rotate
				// passing, because bumping the version is that test's own trigger. Do not
				// delete this as redundant.
				Config: testAccUserConfig_passwordWO_v1_otherTags,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheckTagCount(&user, 1),
					testAccUserConnect("mctest", "wo-secret-one"),
				),
			},
		},
	})
}

func TestAccUser_passwordToPasswordWO(t *testing.T) {
	var user string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy(user),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwordToWO_before,
				Check: resource.ComposeTestCheckFunc(
					testAccUserCheck("rabbitmq_user.test", &user),
					testAccUserConnect("mctest", "foobar"),
					resource.TestCheckResourceAttr("rabbitmq_user.test", "password", "foobar"),
				),
			},
			{
				// Dropping password and adding password_wo + password_wo_version both
				// produce diffs, so one update converges: userPassword prefers password_wo.
				Config: testAccUserConfig_passwordWO_v1,
				Check: resource.ComposeTestCheckFunc(
					testAccUserConnect("mctest", "wo-secret-one"),
					testAccUserCannotConnect("mctest", "foobar"),
					resource.TestCheckResourceAttr("rabbitmq_user.test", "password", ""),
					resource.TestCheckResourceAttr("rabbitmq_user.test", "password_wo_version", "1"),
				),
			},
		},
	})
}

func TestAccUser_passwordValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// The expected messages are the SDK's own ExactlyOneOf/RequiredWith wording, and
		// the key list is alphabetically sorted by the SDK. The attribute-name prefix is
		// deliberately not matched: validation aggregates a diagnostic per attribute, so
		// which one leads is decided by Go map iteration order and is not stable.
		Steps: []resource.TestStep{
			{
				Config:      testAccUserConfig_noPassword,
				ExpectError: regexp.MustCompile("one of `password,password_wo` must be specified"),
			},
			{
				Config:      testAccUserConfig_bothPasswords,
				ExpectError: regexp.MustCompile("only one of `password,password_wo` can be specified"),
			},
			{
				Config:      testAccUserConfig_passwordWO_noVersion,
				ExpectError: regexp.MustCompile("all of `password_wo,password_wo_version` must be specified"),
			},
		},
	})
}

const testAccUserConfig_noPassword = `
resource "rabbitmq_user" "test" {
    name = "mctest"
    tags = ["management"]
}`

const testAccUserConfig_bothPasswords = `
resource "rabbitmq_user" "test" {
    name                = "mctest"
    password            = "foobar"
    password_wo         = "wo-secret-one"
    password_wo_version = 1
    tags                = ["management"]
}`

const testAccUserConfig_passwordWO_noVersion = `
resource "rabbitmq_user" "test" {
    name        = "mctest"
    password_wo = "wo-secret-one"
    tags        = ["management"]
}`

const testAccUserConfig_passwordWO_v1_changedSecret = `
resource "rabbitmq_user" "test" {
    name                = "mctest"
    password_wo         = "wo-secret-ignored"
    password_wo_version = 1
    tags                = ["management"]
}`

const testAccUserConfig_passwordWO_v1_otherTags = `
resource "rabbitmq_user" "test" {
    name                = "mctest"
    password_wo         = "wo-secret-one"
    password_wo_version = 1
    tags                = ["monitoring"]
}`
