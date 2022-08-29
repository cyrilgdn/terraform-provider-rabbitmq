package rabbitmq

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_ENDPOINT", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("Endpoint must not be an empty string"))
					}

					return
				},
			},

			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_USERNAME", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("Username must not be an empty string"))
					}

					return
				},
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_PASSWORD", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("Password must not be an empty string"))
					}

					return
				},
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_INSECURE", nil),
			},

			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_CACERT", ""),
			},

			"clientcert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_CLIENTCERT", ""),
			},

			"clientkey_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_CLIENTKEY", ""),
			},

			"proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RABBITMQ_PROXY", ""),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"rabbitmq_binding":             resourceBinding(),
			"rabbitmq_exchange":            resourceExchange(),
			"rabbitmq_permissions":         resourcePermissions(),
			"rabbitmq_topic_permissions":   resourceTopicPermissions(),
			"rabbitmq_federation_upstream": resourceFederationUpstream(),
			"rabbitmq_operator_policy":     resourceOperatorPolicy(),
			"rabbitmq_policy":              resourcePolicy(),
			"rabbitmq_queue":               resourceQueue(),
			"rabbitmq_user":                resourceUser(),
			"rabbitmq_vhost":               resourceVhost(),
			"rabbitmq_shovel":              resourceShovel(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rabbitmq_exchange": dataSourcesExchange(),
			"rabbitmq_user":     dataSourcesUser(),
			"rabbitmq_vhost":    dataSourcesVhost(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var username = d.Get("username").(string)
	var password = d.Get("password").(string)
	var endpoint = d.Get("endpoint").(string)
	var insecure = d.Get("insecure").(bool)
	var cacertFile = d.Get("cacert_file").(string)
	var clientcertFile = d.Get("clientcert_file").(string)
	var clientkeyFile = d.Get("clientkey_file").(string)
	var proxy = d.Get("proxy").(string)

	// Configure TLS/SSL:
	// Ignore self-signed cert warnings
	// Specify a custom CA / intermediary cert
	// Specify a certificate and key
	tlsConfig := &tls.Config{}
	if cacertFile != "" {
		caCert, err := os.ReadFile(cacertFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	if clientcertFile != "" && clientkeyFile != "" {
		clientPair, err := tls.LoadX509KeyPair(clientcertFile, clientkeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientPair}
	}
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	var proxyURL *url.URL
	if proxy != "" {
		var err error
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", proxy, err)
		}
	}

	// Connect to RabbitMQ management interface
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy: func(req *http.Request) (*url.URL, error) {
			if proxyURL != nil {
				return proxyURL, nil
			}

			return http.ProxyFromEnvironment(req)
		},
	}

	rmqc, err := rabbithole.NewTLSClient(endpoint, username, password, transport)
	if err != nil {
		return nil, err
	}

	return rmqc, nil
}
