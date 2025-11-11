package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccGetGatewayID reads the dynamically created gateway ID from the integration test setup.
// This gateway is created by the integration-test-setup.sh script and points to a running
// MCP time server.
//
// The function will skip the test if:
//   - The gateway ID file doesn't exist (setup script not run)
//   - The file is empty
//
// Returns:
//   - The gateway ID string
func testAccGetGatewayID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-gateway-id.txt")
	if err != nil {
		t.Skipf("Gateway ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	gatewayID := strings.TrimSpace(string(data))
	if gatewayID == "" {
		t.Skip("Gateway ID file is empty - run 'make integration-test-setup' first")
	}

	return gatewayID
}

// TestAccGatewayDataSource_basic tests successful gateway lookup by ID.
// This test verifies that the data source can retrieve a gateway and populate
// all expected attributes.
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//   - Integration test setup completed (creates test gateway)
//
// To run:
//   make testacc-full  # Full lifecycle with setup/teardown
//   make testacc       # Tests only (requires manual setup)
func TestAccGatewayDataSource_basic(t *testing.T) {
	gatewayID := testAccGetGatewayID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccGatewayDataSourceConfig(gatewayID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify lookup attribute
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "id", gatewayID),

					// Verify expected gateway values
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "name", "test-time-server"),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "url", "http://localhost:8002/sse"),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "transport", "SSE"),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "description", "Test gateway for integration tests"),

					// Verify boolean attributes
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "reachable"),

					// Verify timestamp attributes are populated
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "updated_at"),
				),
			},
		},
	})
}

// TestAccGatewayDataSource_missingID tests error handling when ID is not provided.
// The data source should return a clear error message when the required ID attribute
// is missing from the configuration.
//
// To run:
//   TF_ACC=1 go test -v ./internal/data/ -run TestAccGatewayDataSource_missingID
func TestAccGatewayDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGatewayDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

// TestAccGatewayDataSource_nonExistent tests error handling for non-existent gateway ID.
// The data source should return an appropriate error when attempting to look up
// a gateway that doesn't exist.
//
// To run:
//   TF_ACC=1 go test -v ./internal/data/ -run TestAccGatewayDataSource_nonExistent
func TestAccGatewayDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGatewayDataSourceConfig("non-existent-gateway-id-12345"),
				ExpectError: regexp.MustCompile(`Failed to Read Gateway|Unable to read gateway`),
			},
		},
	})
}

// TestAccGatewayDataSource_allAttributes tests that all gateway attributes are
// properly mapped and accessible. This test verifies optional/nullable fields
// are handled correctly.
//
// To run:
//   make testacc-full  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/data/ -run TestAccGatewayDataSource_allAttributes
func TestAccGatewayDataSource_allAttributes(t *testing.T) {
	gatewayID := testAccGetGatewayID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayDataSourceConfig(gatewayID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Core fields
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "id", gatewayID),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "name", "test-time-server"),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "url", "http://localhost:8002/sse"),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "transport", "SSE"),
					resource.TestCheckResourceAttr("data.contextforge_gateway.test", "description", "Test gateway for integration tests"),
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "reachable"),

					// Timestamps
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_gateway.test", "updated_at"),

					// Note: Other attributes (auth fields, tags, metadata) may be null/empty
					// depending on the gateway configuration. We don't check them with TestCheckResourceAttrSet
					// as that would fail for null values.
				),
			},
		},
	})
}

// testAccGatewayDataSourceConfig returns the Terraform configuration for gateway lookup by ID.
// This helper function generates HCL configuration for testing the gateway data source.
//
// Parameters:
//   - gatewayID: The ID of the gateway to look up
//
// Returns:
//   - HCL configuration string with the data source definition
func testAccGatewayDataSourceConfig(gatewayID string) string {
	return fmt.Sprintf(`
data "contextforge_gateway" "test" {
  id = %[1]q
}
`, gatewayID)
}

// testAccGatewayDataSourceConfigMissingID returns invalid Terraform configuration
// with missing required ID attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required id attribute
func testAccGatewayDataSourceConfigMissingID() string {
	return `
data "contextforge_gateway" "test" {
  # Missing required id attribute
}
`
}

// testAccGatewayDataSourceConfigWithOutputs returns Terraform configuration
// that includes outputs to verify data source attribute values.
//
// This helper is useful for debugging and verifying that specific attribute
// values are correctly retrieved from the API.
//
// Parameters:
//   - gatewayID: The ID of the gateway to look up
//
// Returns:
//   - HCL configuration string with data source and outputs
func testAccGatewayDataSourceConfigWithOutputs(gatewayID string) string {
	return fmt.Sprintf(`
data "contextforge_gateway" "test" {
  id = %[1]q
}

output "gateway_name" {
  value = data.contextforge_gateway.test.name
}

output "gateway_url" {
  value = data.contextforge_gateway.test.url
}

output "gateway_enabled" {
  value = data.contextforge_gateway.test.enabled
}
`, gatewayID)
}
