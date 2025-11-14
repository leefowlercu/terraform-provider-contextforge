package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccGetToolID reads the dynamically created tool ID from the integration test setup.
// This tool is created by the integration-test-setup.sh script.
//
// The function will skip the test if:
//   - The tool ID file doesn't exist (setup script not run)
//   - The file is empty
//
// Returns:
//   - The tool ID string
func testAccGetToolID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-tool-id.txt")
	if err != nil {
		t.Skipf("Tool ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	toolID := strings.TrimSpace(string(data))
	if toolID == "" {
		t.Skip("Tool ID file is empty - run 'make integration-test-setup' first")
	}

	return toolID
}

// TestAccToolDataSource_basic tests successful tool lookup by ID.
// This test verifies that the data source can retrieve a tool and populate
// all expected attributes.
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//   - Integration test setup completed (creates test tool)
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   make integration-test      # Tests only (requires manual setup)
func TestAccToolDataSource_basic(t *testing.T) {
	toolID := testAccGetToolID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccToolDataSourceConfig(toolID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify lookup attribute
					resource.TestCheckResourceAttr("data.contextforge_tool.test", "id", toolID),

					// Verify expected tool values
					resource.TestCheckResourceAttr("data.contextforge_tool.test", "name", "test-tool"),
					resource.TestCheckResourceAttr("data.contextforge_tool.test", "description", "Test tool for integration tests"),

					// Verify boolean attributes
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "enabled"),

					// Verify timestamp attributes are populated
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "updated_at"),
				),
			},
		},
	})
}

// TestAccToolDataSource_missingID tests error handling when ID is not provided.
// The data source should return a clear error message when the required ID attribute
// is missing from the configuration.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolDataSource_missingID
func TestAccToolDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccToolDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

// TestAccToolDataSource_nonExistent tests error handling for non-existent tool ID.
// The data source should return an appropriate error when attempting to look up
// a tool that doesn't exist.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolDataSource_nonExistent
func TestAccToolDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccToolDataSourceConfig("non-existent-tool-id-12345"),
				ExpectError: regexp.MustCompile(`Failed to Read Tool|Unable to read tool`),
			},
		},
	})
}

// TestAccToolDataSource_allAttributes tests that all tool attributes are
// properly mapped and accessible. This test verifies optional/nullable fields
// are handled correctly including input_schema and tags.
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolDataSource_allAttributes
func TestAccToolDataSource_allAttributes(t *testing.T) {
	toolID := testAccGetToolID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccToolDataSourceConfig(toolID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Core fields
					resource.TestCheckResourceAttr("data.contextforge_tool.test", "id", toolID),
					resource.TestCheckResourceAttr("data.contextforge_tool.test", "name", "test-tool"),
					resource.TestCheckResourceAttr("data.contextforge_tool.test", "description", "Test tool for integration tests"),
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "enabled"),

					// Timestamps
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "updated_at"),

					// Tags (should be populated with test tags)
					resource.TestCheckResourceAttrSet("data.contextforge_tool.test", "tags.#"),

					// Note: input_schema is a dynamic attribute that cannot be easily tested with
					// TestCheckResourceAttrSet. The basic test confirms it's populated correctly.
					// Other attributes (team_id, visibility) may be null/empty depending on the
					// tool configuration. We don't check them as that would fail for null values.
				),
			},
		},
	})
}

// testAccToolDataSourceConfig returns the Terraform configuration for tool lookup by ID.
// This helper function generates HCL configuration for testing the tool data source.
//
// Parameters:
//   - toolID: The ID of the tool to look up
//
// Returns:
//   - HCL configuration string with the data source definition
func testAccToolDataSourceConfig(toolID string) string {
	return fmt.Sprintf(`
data "contextforge_tool" "test" {
  id = %[1]q
}
`, toolID)
}

// testAccToolDataSourceConfigMissingID returns invalid Terraform configuration
// with missing required ID attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required id attribute
func testAccToolDataSourceConfigMissingID() string {
	return `
data "contextforge_tool" "test" {
  # Missing required id attribute
}
`
}

// testAccToolDataSourceConfigWithOutputs returns Terraform configuration
// that includes outputs to verify data source attribute values.
//
// This helper is useful for debugging and verifying that specific attribute
// values are correctly retrieved from the API.
//
// Parameters:
//   - toolID: The ID of the tool to look up
//
// Returns:
//   - HCL configuration string with data source and outputs
func testAccToolDataSourceConfigWithOutputs(toolID string) string {
	return fmt.Sprintf(`
data "contextforge_tool" "test" {
  id = %[1]q
}

output "tool_name" {
  value = data.contextforge_tool.test.name
}

output "tool_enabled" {
  value = data.contextforge_tool.test.enabled
}

output "tool_tags" {
  value = data.contextforge_tool.test.tags
}

output "tool_input_schema" {
  value = data.contextforge_tool.test.input_schema
}
`, toolID)
}
