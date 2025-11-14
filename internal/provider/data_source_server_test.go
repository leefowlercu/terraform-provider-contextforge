package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccGetServerID reads the dynamically created server ID from the integration test setup.
// This server is created by the integration-test-setup.sh script.
//
// The function will skip the test if:
//   - The server ID file doesn't exist (setup script not run)
//   - The file is empty
//
// Returns:
//   - The server ID string
func testAccGetServerID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-server-id.txt")
	if err != nil {
		t.Skipf("Server ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	serverID := strings.TrimSpace(string(data))
	if serverID == "" {
		t.Skip("Server ID file is empty - run 'make integration-test-setup' first")
	}

	return serverID
}

// TestAccServerDataSource_basic tests successful server lookup by ID.
// This test verifies that the data source can retrieve a server and populate
// all expected attributes.
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//   - Integration test setup completed (creates test server)
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   make integration-test      # Tests only (requires manual setup)
func TestAccServerDataSource_basic(t *testing.T) {
	serverID := testAccGetServerID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccServerDataSourceConfig(serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify lookup attribute
					resource.TestCheckResourceAttr("data.contextforge_server.test", "id", serverID),

					// Verify expected server values
					resource.TestCheckResourceAttr("data.contextforge_server.test", "name", "test-server"),
					resource.TestCheckResourceAttr("data.contextforge_server.test", "description", "Test server for integration tests"),

					// Verify boolean attributes
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "is_active"),

					// Verify timestamp attributes are populated
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "updated_at"),
				),
			},
		},
	})
}

// TestAccServerDataSource_missingID tests error handling when ID is not provided.
// The data source should return a clear error message when the required ID attribute
// is missing from the configuration.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccServerDataSource_missingID
func TestAccServerDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServerDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

// TestAccServerDataSource_nonExistent tests error handling for non-existent server ID.
// The data source should return an appropriate error when attempting to look up
// a server that doesn't exist.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccServerDataSource_nonExistent
func TestAccServerDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServerDataSourceConfig("non-existent-server-id-12345"),
				ExpectError: regexp.MustCompile(`Failed to Read Server|Unable to read server`),
			},
		},
	})
}

// TestAccServerDataSource_allAttributes tests that all server attributes are
// properly mapped and accessible. This test verifies optional/nullable fields
// are handled correctly including nested metrics, association arrays, and tags.
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccServerDataSource_allAttributes
func TestAccServerDataSource_allAttributes(t *testing.T) {
	serverID := testAccGetServerID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig(serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Core fields
					resource.TestCheckResourceAttr("data.contextforge_server.test", "id", serverID),
					resource.TestCheckResourceAttr("data.contextforge_server.test", "name", "test-server"),
					resource.TestCheckResourceAttr("data.contextforge_server.test", "description", "Test server for integration tests"),
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "is_active"),

					// Timestamps
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "updated_at"),

					// Tags (should be populated with test tags)
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "tags.#"),

					// Metrics (should be present as nested object)
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "metrics.total_executions"),
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "metrics.successful_executions"),
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "metrics.failed_executions"),
					resource.TestCheckResourceAttrSet("data.contextforge_server.test", "metrics.failure_rate"),

					// Note: Other attributes (association fields, icon, organizational metadata)
					// may be null/empty depending on the server configuration.
					// We don't check them with TestCheckResourceAttrSet as that would fail for null values.
				),
			},
		},
	})
}

// testAccServerDataSourceConfig returns the Terraform configuration for server lookup by ID.
// This helper function generates HCL configuration for testing the server data source.
//
// Parameters:
//   - serverID: The ID of the server to look up
//
// Returns:
//   - HCL configuration string with the data source definition
func testAccServerDataSourceConfig(serverID string) string {
	return fmt.Sprintf(`
data "contextforge_server" "test" {
  id = %[1]q
}
`, serverID)
}

// testAccServerDataSourceConfigMissingID returns invalid Terraform configuration
// with missing required ID attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required id attribute
func testAccServerDataSourceConfigMissingID() string {
	return `
data "contextforge_server" "test" {
  # Missing required id attribute
}
`
}

// testAccServerDataSourceConfigWithOutputs returns Terraform configuration
// that includes outputs to verify data source attribute values.
//
// This helper is useful for debugging and verifying that specific attribute
// values are correctly retrieved from the API.
//
// Parameters:
//   - serverID: The ID of the server to look up
//
// Returns:
//   - HCL configuration string with data source and outputs
func testAccServerDataSourceConfigWithOutputs(serverID string) string {
	return fmt.Sprintf(`
data "contextforge_server" "test" {
  id = %[1]q
}

output "server_name" {
  value = data.contextforge_server.test.name
}

output "server_is_active" {
  value = data.contextforge_server.test.is_active
}

output "server_tags" {
  value = data.contextforge_server.test.tags
}

output "server_metrics_total_executions" {
  value = data.contextforge_server.test.metrics.total_executions
}
`, serverID)
}
