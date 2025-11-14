package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccGetResourceID reads the dynamically created resource ID from the integration test setup.
// This resource is created by the integration-test-setup.sh script.
//
// The function will skip the test if:
//   - The resource ID file doesn't exist (setup script not run)
//   - The file is empty
//
// Returns:
//   - The resource ID string
func testAccGetResourceID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-resource-id.txt")
	if err != nil {
		t.Skipf("Resource ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	resourceID := strings.TrimSpace(string(data))
	if resourceID == "" {
		t.Skip("Resource ID file is empty - run 'make integration-test-setup' first")
	}

	return resourceID
}

// TestAccResourceDataSource_basic tests successful resource lookup by ID.
// This test verifies that the data source can retrieve a resource and populate
// all expected attributes.
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//   - Integration test setup completed (creates test resource)
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   make integration-test      # Tests only (requires manual setup)
func TestAccResourceDataSource_basic(t *testing.T) {
	resourceID := testAccGetResourceID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccResourceDataSourceConfig(resourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify lookup attribute
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "id", resourceID),

					// Verify expected resource values
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "name", "test-resource"),
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "description", "Test resource for integration tests"),
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "uri", "test://integration/resource"),

					// Verify boolean attributes
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "is_active"),

					// Verify timestamp attributes are populated
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "updated_at"),
				),
			},
		},
	})
}

// TestAccResourceDataSource_missingID tests error handling when ID is not provided.
// The data source should return a clear error message when the required ID attribute
// is missing from the configuration.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceDataSource_missingID
func TestAccResourceDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

// TestAccResourceDataSource_nonExistent tests error handling for non-existent resource ID.
// The data source should return an appropriate error when attempting to look up
// a resource that doesn't exist.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceDataSource_nonExistent
func TestAccResourceDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceDataSourceConfig("999999"),
				ExpectError: regexp.MustCompile(`Resource Not Found|Unable to find resource`),
			},
		},
	})
}

// TestAccResourceDataSource_allAttributes tests that all resource attributes are
// properly mapped and accessible. This test verifies optional/nullable fields
// are handled correctly including nested metrics, mime_type, size, and tags.
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceDataSource_allAttributes
func TestAccResourceDataSource_allAttributes(t *testing.T) {
	resourceID := testAccGetResourceID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSourceConfig(resourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Core fields
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "id", resourceID),
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "name", "test-resource"),
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "description", "Test resource for integration tests"),
					resource.TestCheckResourceAttr("data.contextforge_resource.test", "uri", "test://integration/resource"),
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "is_active"),

					// Timestamps
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "updated_at"),

					// Tags (should be populated with test tags)
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "tags.#"),

					// Metrics (should be present as nested object)
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "metrics.total_executions"),
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "metrics.successful_executions"),
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "metrics.failed_executions"),
					resource.TestCheckResourceAttrSet("data.contextforge_resource.test", "metrics.failure_rate"),

					// Note: Other attributes (mime_type, size, organizational metadata)
					// may be null/empty depending on the resource configuration.
					// We don't check them with TestCheckResourceAttrSet as that would fail for null values.
				),
			},
		},
	})
}

// testAccResourceDataSourceConfig returns the Terraform configuration for resource lookup by ID.
// This helper function generates HCL configuration for testing the resource data source.
//
// Parameters:
//   - resourceID: The ID of the resource to look up
//
// Returns:
//   - HCL configuration string with the data source definition
func testAccResourceDataSourceConfig(resourceID string) string {
	return fmt.Sprintf(`
data "contextforge_resource" "test" {
  id = %[1]q
}
`, resourceID)
}

// testAccResourceDataSourceConfigMissingID returns invalid Terraform configuration
// with missing required ID attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required id attribute
func testAccResourceDataSourceConfigMissingID() string {
	return `
data "contextforge_resource" "test" {
  # Missing required id attribute
}
`
}

// testAccResourceDataSourceConfigWithOutputs returns Terraform configuration
// that includes outputs to verify data source attribute values.
//
// This helper is useful for debugging and verifying that specific attribute
// values are correctly retrieved from the API.
//
// Parameters:
//   - resourceID: The ID of the resource to look up
//
// Returns:
//   - HCL configuration string with data source and outputs
func testAccResourceDataSourceConfigWithOutputs(resourceID string) string {
	return fmt.Sprintf(`
data "contextforge_resource" "test" {
  id = %[1]q
}

output "resource_name" {
  value = data.contextforge_resource.test.name
}

output "resource_uri" {
  value = data.contextforge_resource.test.uri
}

output "resource_is_active" {
  value = data.contextforge_resource.test.is_active
}

output "resource_tags" {
  value = data.contextforge_resource.test.tags
}

output "resource_metrics_total_executions" {
  value = data.contextforge_resource.test.metrics.total_executions
}
`, resourceID)
}
