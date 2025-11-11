package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"contextforge": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables are set for
// acceptance tests. This function should be called in the PreCheck field of
// every resource.TestCase.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("CONTEXTFORGE_ADDR"); v == "" {
		t.Fatal("CONTEXTFORGE_ADDR must be set for acceptance tests")
	}
	if v := os.Getenv("CONTEXTFORGE_TOKEN"); v == "" {
		t.Fatal("CONTEXTFORGE_TOKEN must be set for acceptance tests")
	}
}
