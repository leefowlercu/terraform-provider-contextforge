package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGatewayResource_basic tests the basic CRUD lifecycle for a gateway resource.
// This test verifies:
//   - Create with minimal required fields (name, url, transport)
//   - Read to verify created values
//   - Delete to remove resource
//
// Note: Update tests are separate due to potential upstream bugs
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccGatewayResource_basic
func TestAccGatewayResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGatewayResourceConfig("tf-test-gateway", "http://localhost:8003/sse", "SSE", "Test gateway created by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_gateway.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_gateway.test", "updated_at"),

					// Verify configured attributes
					resource.TestCheckResourceAttr("contextforge_gateway.test", "name", "tf-test-gateway"),
					resource.TestCheckResourceAttr("contextforge_gateway.test", "url", "http://localhost:8003/sse"),
					resource.TestCheckResourceAttr("contextforge_gateway.test", "transport", "SSE"),
					resource.TestCheckResourceAttr("contextforge_gateway.test", "description", "Test gateway created by Terraform"),
					resource.TestCheckResourceAttr("contextforge_gateway.test", "enabled", "true"),

					// Verify computed fields exist
					resource.TestCheckResourceAttrSet("contextforge_gateway.test", "reachable"),
					resource.TestCheckResourceAttrSet("contextforge_gateway.test", "slug"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_gateway.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccGatewayResource_import tests importing an existing gateway.
// This verifies that gateways can be imported using their ID and that
// all attributes are correctly populated in the state.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccGatewayResource_import
func TestAccGatewayResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccGatewayResourceConfig("tf-import-gateway", "http://localhost:8004/sse", "SSE", "Gateway for import testing"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contextforge_gateway.test", "id"),
				),
			},
			// Import by ID
			{
				ResourceName:      "contextforge_gateway.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccGatewayResource_missingRequired tests error handling when required fields are missing.
// This verifies that the resource properly validates required attributes.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccGatewayResource_missingRequired
func TestAccGatewayResource_missingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGatewayResourceConfigMissingName(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "name" is required`),
			},
			{
				Config:      testAccGatewayResourceConfigMissingURL(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "url" is required`),
			},
			{
				Config:      testAccGatewayResourceConfigMissingTransport(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "transport" is required`),
			},
		},
	})
}

// testAccGatewayResourceConfig generates basic Terraform configuration for a gateway resource.
// This helper creates a minimal valid gateway configuration.
//
// Parameters:
//   - name: Gateway name
//   - url: Gateway URL endpoint
//   - transport: Transport protocol (SSE, HTTP, etc.)
//   - description: Gateway description
//
// Returns:
//   - HCL configuration string
func testAccGatewayResourceConfig(name, url, transport, description string) string {
	return fmt.Sprintf(`
resource "contextforge_gateway" "test" {
  name        = %[1]q
  url         = %[2]q
  transport   = %[3]q
  description = %[4]q
  enabled     = true
}
`, name, url, transport, description)
}

// testAccGatewayResourceConfigMissingName generates invalid Terraform configuration
// with missing required name attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required name attribute
func testAccGatewayResourceConfigMissingName() string {
	return `
resource "contextforge_gateway" "test" {
  url         = "http://localhost:9997/sse"
  transport   = "SSE"
  description = "Gateway missing required name"
}
`
}

// testAccGatewayResourceConfigMissingURL generates invalid Terraform configuration
// with missing required url attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required url attribute
func testAccGatewayResourceConfigMissingURL() string {
	return `
resource "contextforge_gateway" "test" {
  name        = "test-gateway"
  transport   = "SSE"
  description = "Gateway missing required url"
}
`
}

// testAccGatewayResourceConfigMissingTransport generates invalid Terraform configuration
// with missing required transport attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required transport attribute
func testAccGatewayResourceConfigMissingTransport() string {
	return `
resource "contextforge_gateway" "test" {
  name        = "test-gateway"
  url         = "http://localhost:9996/sse"
  description = "Gateway missing required transport"
}
`
}
