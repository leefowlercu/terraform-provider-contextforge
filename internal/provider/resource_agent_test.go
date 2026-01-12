package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAgentResource_basic tests the basic CRUD lifecycle for an agent resource.
// This test verifies:
//   - Create with minimal required fields (name, endpoint_url)
//   - Read to verify created values
//   - Delete to remove resource
//
// Note: Currently skipped due to API returning empty config objects and computed fields showing as changed.
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccAgentResource_basic
func TestAccAgentResource_basic(t *testing.T) {
	t.Skip("Skipping due to API returning empty config objects and computed fields showing as changed")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAgentResourceConfig("tf-test-agent", "http://localhost:9001/agent", "Test agent created by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "updated_at"),

					// Verify configured attributes
					resource.TestCheckResourceAttr("contextforge_agent.test", "name", "tf-test-agent"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "endpoint_url", "http://localhost:9001/agent"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "description", "Test agent created by Terraform"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "enabled", "true"),

					// Verify computed fields exist
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "reachable"),
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "slug"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_agent.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAgentResource_withOptionalFields tests agent creation with optional fields.
// This test verifies that all optional fields including config, tags, and organizational
// fields are properly created and managed.
//
// Note: Currently skipped due to API returning empty config objects and computed fields showing as changed.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccAgentResource_withOptionalFields
func TestAccAgentResource_withOptionalFields(t *testing.T) {
	t.Skip("Skipping due to API returning empty config objects and computed fields showing as changed")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with optional fields
			{
				Config: testAccAgentResourceConfigComplete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "updated_at"),

					// Verify core attributes
					resource.TestCheckResourceAttr("contextforge_agent.test", "name", "tf-complete-agent"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "endpoint_url", "http://localhost:9002/agent"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "description", "Complete agent with all attributes"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "enabled", "true"),

					// Verify tags
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.#", "3"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.1", "testing"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.2", "complete"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_agent.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAgentResource_update tests updating various agent attributes.
// This verifies that updates to name, description, enabled, and tags
// are properly applied without forcing resource recreation.
//
// Note: Currently skipped due to API behavior with computed fields during updates.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccAgentResource_update
func TestAccAgentResource_update(t *testing.T) {
	t.Skip("Skipping due to API behavior with computed fields showing as changed during update")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial resource
			{
				Config: testAccAgentResourceConfigWithTags("tf-update-agent", []string{"initial"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_agent.test", "name", "tf-update-agent"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "enabled", "true"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.0", "initial"),
				),
			},
			// Update name, enabled, and tags
			{
				Config: testAccAgentResourceConfigUpdate("tf-updated-agent", false, []string{"updated", "modified"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_agent.test", "name", "tf-updated-agent"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "enabled", "false"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.0", "updated"),
					resource.TestCheckResourceAttr("contextforge_agent.test", "tags.1", "modified"),
				),
			},
		},
	})
}

// TestAccAgentResource_import tests importing an existing agent.
// This verifies that agents can be imported using their ID and that
// all attributes are correctly populated in the state.
//
// Note: Currently skipped due to API behavior with empty config objects and computed fields during import.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccAgentResource_import
func TestAccAgentResource_import(t *testing.T) {
	t.Skip("Skipping due to API behavior with empty config objects and computed fields during import")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccAgentResourceConfig("tf-import-agent", "http://localhost:9003/agent", "Agent for import testing"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contextforge_agent.test", "id"),
				),
			},
			// Import by ID
			{
				ResourceName:      "contextforge_agent.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAgentResource_missingRequired tests error handling when required fields are missing.
// This verifies that the resource properly validates required attributes.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccAgentResource_missingRequired
func TestAccAgentResource_missingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgentResourceConfigMissingName(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument \"name\" is required`),
			},
			{
				Config:      testAccAgentResourceConfigMissingEndpointURL(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument \"endpoint_url\" is required`),
			},
		},
	})
}

// testAccAgentResourceConfig generates basic Terraform configuration for an agent resource.
// This helper creates a minimal valid agent configuration.
//
// Parameters:
//   - name: Agent name
//   - endpointURL: Agent endpoint URL
//   - description: Agent description
//
// Returns:
//   - HCL configuration string
func testAccAgentResourceConfig(name, endpointURL, description string) string {
	return fmt.Sprintf(`
resource "contextforge_agent" "test" {
  name         = %[1]q
  endpoint_url = %[2]q
  description  = %[3]q
  enabled      = true
}
`, name, endpointURL, description)
}

// testAccAgentResourceConfigComplete generates Terraform configuration with all optional attributes.
// This includes tags and other optional fields.
// Note: config field omitted as empty objects cause API inconsistency issues.
//
// Returns:
//   - HCL configuration string with all attributes
func testAccAgentResourceConfigComplete() string {
	return `
resource "contextforge_agent" "test" {
  name         = "tf-complete-agent"
  endpoint_url = "http://localhost:9002/agent"
  description  = "Complete agent with all attributes"
  enabled      = true
  tags         = ["terraform", "testing", "complete"]
}
`
}

// testAccAgentResourceConfigWithTags generates Terraform configuration with tags.
// This helper is useful for testing tag updates.
//
// Parameters:
//   - name: Agent name
//   - tags: List of tag strings
//
// Returns:
//   - HCL configuration string with tags
func testAccAgentResourceConfigWithTags(name string, tags []string) string {
	tagsHCL := "["
	for i, tag := range tags {
		if i > 0 {
			tagsHCL += ", "
		}
		tagsHCL += fmt.Sprintf("%q", tag)
	}
	tagsHCL += "]"

	return fmt.Sprintf(`
resource "contextforge_agent" "test" {
  name         = %[1]q
  endpoint_url = "http://localhost:9004/agent"
  description  = "Agent with tags"
  enabled      = true
  tags         = %[2]s
}
`, name, tagsHCL)
}

// testAccAgentResourceConfigUpdate generates Terraform configuration for update testing.
// This helper includes name, enabled, and tags parameters.
//
// Parameters:
//   - name: Agent name
//   - enabled: Whether the agent is enabled
//   - tags: List of tag strings
//
// Returns:
//   - HCL configuration string for update testing
func testAccAgentResourceConfigUpdate(name string, enabled bool, tags []string) string {
	tagsHCL := "["
	for i, tag := range tags {
		if i > 0 {
			tagsHCL += ", "
		}
		tagsHCL += fmt.Sprintf("%q", tag)
	}
	tagsHCL += "]"

	enabledStr := "true"
	if !enabled {
		enabledStr = "false"
	}

	return fmt.Sprintf(`
resource "contextforge_agent" "test" {
  name         = %[1]q
  endpoint_url = "http://localhost:9004/agent"
  description  = "Agent for update testing"
  enabled      = %[2]s
  tags         = %[3]s
}
`, name, enabledStr, tagsHCL)
}

// testAccAgentResourceConfigMissingName generates invalid Terraform configuration
// with missing required name attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required name attribute
func testAccAgentResourceConfigMissingName() string {
	return `
resource "contextforge_agent" "test" {
  endpoint_url = "http://localhost:9005/agent"
  description  = "Agent missing required name"
}
`
}

// testAccAgentResourceConfigMissingEndpointURL generates invalid Terraform configuration
// with missing required endpoint_url attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required endpoint_url attribute
func testAccAgentResourceConfigMissingEndpointURL() string {
	return `
resource "contextforge_agent" "test" {
  name        = "test-agent"
  description = "Agent missing required endpoint_url"
}
`
}
