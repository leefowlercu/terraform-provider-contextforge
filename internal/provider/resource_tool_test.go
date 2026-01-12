package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccToolResource_basic tests the full CRUD lifecycle for a tool resource.
// This test verifies:
//   - Create with minimal required fields
//   - Read to verify created values
//   - Update to modify fields
//   - Delete to remove resource
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolResource_basic
func TestAccToolResource_basic(t *testing.T) {
	t.Skip("Skipping due to upstream ContextForge bug - tool update API does not update name field. See docs/upstream-bugs/contextforge-tool-update-bug.md")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccToolResourceConfig("tf-test-tool", "Test tool created by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "updated_at"),

					// Verify configured attributes
					resource.TestCheckResourceAttr("contextforge_tool.test", "name", "tf-test-tool"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "description", "Test tool created by Terraform"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "true"),
				),
			},
			// Update and Read testing
			{
				Config: testAccToolResourceConfig("tf-test-tool-updated", "Updated tool description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID remains the same
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "id"),

					// Verify updated attributes
					resource.TestCheckResourceAttr("contextforge_tool.test", "name", "tf-test-tool-updated"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "description", "Updated tool description"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_tool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccToolResource_complete tests tool creation with all optional attributes.
// This test verifies that all fields including input_schema, tags, and organizational
// fields are properly created and managed.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolResource_complete
func TestAccToolResource_complete(t *testing.T) {
	t.Skip("Skipping due to upstream ContextForge bug - API returns default empty input_schema causing drift. See docs/upstream-bugs/contextforge-tool-update-bug.md")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				Config: testAccToolResourceConfigComplete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "updated_at"),

					// Verify core attributes
					resource.TestCheckResourceAttr("contextforge_tool.test", "name", "tf-complete-tool"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "description", "Complete tool with all attributes"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "true"),

					// Verify organizational attributes
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.#", "3"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.1", "testing"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.2", "complete"),

					// Note: input_schema is a Dynamic type - check if it exists
					// Dynamic attributes cannot be checked with TestCheckResourceAttrSet
					// resource.TestCheckResourceAttrSet("contextforge_tool.test", "input_schema"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_tool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccToolResource_update tests updating various tool attributes.
// This verifies that updates to name, description, enabled flag, and tags
// are properly applied without forcing resource recreation.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolResource_update
func TestAccToolResource_update(t *testing.T) {
	t.Skip("Skipping due to upstream ContextForge bug - tool update API does not update name field. See docs/upstream-bugs/contextforge-tool-update-bug.md")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial resource
			{
				Config: testAccToolResourceConfigWithTags("tf-update-tool", []string{"initial"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_tool.test", "name", "tf-update-tool"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.0", "initial"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "true"),
				),
			},
			// Update name and tags
			{
				Config: testAccToolResourceConfigWithTags("tf-updated-tool", []string{"updated", "modified"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_tool.test", "name", "tf-updated-tool"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.0", "updated"),
					resource.TestCheckResourceAttr("contextforge_tool.test", "tags.1", "modified"),
				),
			},
			// Update enabled flag
			{
				Config: testAccToolResourceConfigDisabled("tf-updated-tool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "false"),
				),
			},
		},
	})
}

// TestAccToolResource_enabledToggle tests toggling the enabled flag.
// This verifies that the enabled attribute can be changed between true and false
// without forcing resource recreation.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolResource_enabledToggle
func TestAccToolResource_enabledToggle(t *testing.T) {
	t.Skip("Skipping due to upstream ContextForge bug - tool update API does not update enabled field. See docs/upstream-bugs/contextforge-tool-update-bug.md")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create enabled tool
			{
				Config: testAccToolResourceConfig("tf-toggle-tool", "Tool for testing enabled toggle"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "true"),
				),
			},
			// Disable tool
			{
				Config: testAccToolResourceConfigDisabled("tf-toggle-tool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "false"),
				),
			},
			// Re-enable tool
			{
				Config: testAccToolResourceConfig("tf-toggle-tool", "Tool for testing enabled toggle"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_tool.test", "enabled", "true"),
				),
			},
		},
	})
}

// TestAccToolResource_import tests importing an existing tool.
// This verifies that tools can be imported using their ID and that
// all attributes are correctly populated in the state.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolResource_import
func TestAccToolResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccToolResourceConfig("tf-import-tool", "Tool for import testing"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contextforge_tool.test", "id"),
				),
			},
			// Import by ID
			{
				ResourceName:      "contextforge_tool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccToolResource_missingRequired tests error handling when required fields are missing.
// This verifies that the resource properly validates required attributes.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccToolResource_missingRequired
func TestAccToolResource_missingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccToolResourceConfigMissingRequired(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "name" is required`),
			},
		},
	})
}

// testAccToolResourceConfig generates basic Terraform configuration for a tool resource.
// This helper creates a minimal valid tool configuration.
//
// Parameters:
//   - name: Tool name
//   - description: Tool description
//
// Returns:
//   - HCL configuration string
func testAccToolResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "contextforge_tool" "test" {
  name        = %[1]q
  description = %[2]q
  enabled     = true
}
`, name, description)
}

// testAccToolResourceConfigComplete generates Terraform configuration with all attributes.
// This includes input_schema as a Dynamic type and all optional fields.
//
// Returns:
//   - HCL configuration string with all attributes
func testAccToolResourceConfigComplete() string {
	return `
resource "contextforge_tool" "test" {
  name        = "tf-complete-tool"
  description = "Complete tool with all attributes"
  enabled     = true
  tags        = ["terraform", "testing", "complete"]

  input_schema = {
    type = "object"
    properties = {
      message = {
        type        = "string"
        description = "The message to process"
      }
      priority = {
        type    = "integer"
        minimum = 1
        maximum = 10
      }
    }
    required = ["message"]
  }
}
`
}

// testAccToolResourceConfigWithTags generates Terraform configuration with tags.
// This helper is useful for testing tag updates.
//
// Parameters:
//   - name: Tool name
//   - tags: List of tag strings
//
// Returns:
//   - HCL configuration string with tags
func testAccToolResourceConfigWithTags(name string, tags []string) string {
	tagsHCL := "["
	for i, tag := range tags {
		if i > 0 {
			tagsHCL += ", "
		}
		tagsHCL += fmt.Sprintf("%q", tag)
	}
	tagsHCL += "]"

	return fmt.Sprintf(`
resource "contextforge_tool" "test" {
  name        = %[1]q
  description = "Tool with tags"
  enabled     = true
  tags        = %[2]s
}
`, name, tagsHCL)
}

// testAccToolResourceConfigDisabled generates Terraform configuration with enabled = false.
// This helper is useful for testing the enabled toggle functionality.
//
// Parameters:
//   - name: Tool name
//
// Returns:
//   - HCL configuration string with enabled = false
func testAccToolResourceConfigDisabled(name string) string {
	return fmt.Sprintf(`
resource "contextforge_tool" "test" {
  name        = %[1]q
  description = "Disabled tool"
  enabled     = false
}
`, name)
}

// testAccToolResourceConfigMissingRequired generates invalid Terraform configuration
// with missing required name attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required name attribute
func testAccToolResourceConfigMissingRequired() string {
	return `
resource "contextforge_tool" "test" {
  description = "Tool missing required name"
}
`
}

// testAccToolResourceConfigWithSchema generates Terraform configuration with a custom input_schema.
// This helper is useful for testing Dynamic type handling.
//
// Parameters:
//   - name: Tool name
//   - schemaJSON: JSON Schema as a map
//
// Returns:
//   - HCL configuration string with custom input_schema
func testAccToolResourceConfigWithSchema(name string) string {
	return fmt.Sprintf(`
resource "contextforge_tool" "test" {
  name        = %[1]q
  description = "Tool with custom schema"
  enabled     = true

  input_schema = {
    type = "object"
    properties = {
      action = {
        type = "string"
        enum = ["create", "update", "delete"]
      }
      resource_id = {
        type = "string"
      }
    }
    required = ["action", "resource_id"]
  }
}
`, name)
}
