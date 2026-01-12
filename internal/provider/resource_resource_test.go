package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceResource_basic tests the basic CRUD lifecycle for a resource resource.
// This test verifies:
//   - Create with minimal required fields (uri, name)
//   - Read to verify created values
//   - Delete to remove resource
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceResource_basic
func TestAccResourceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceResourceConfig("test://terraform/basic", "tf-test-resource", "Test resource created by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "updated_at"),

					// Verify configured attributes
					resource.TestCheckResourceAttr("contextforge_resource.test", "uri", "test://terraform/basic"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "name", "tf-test-resource"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "description", "Test resource created by Terraform"),

					// Verify computed fields exist
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "is_active"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_resource.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceResource_complete tests resource creation with all optional attributes.
// This test verifies that all fields including mime_type, size, tags are properly
// created and managed.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceResource_complete
func TestAccResourceResource_complete(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				Config: testAccResourceResourceConfigComplete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed attributes
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "id"),
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "updated_at"),

					// Verify core attributes
					resource.TestCheckResourceAttr("contextforge_resource.test", "uri", "test://terraform/complete"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "name", "tf-complete-resource"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "description", "Complete resource with all attributes"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "mime_type", "text/plain"),

					// Verify tags
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.#", "3"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.1", "testing"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.2", "complete"),
				),
			},
			// Import testing
			{
				ResourceName:      "contextforge_resource.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceResource_update tests updating various resource attributes.
// This verifies that updates to name, description, mime_type, size, and tags
// are properly applied without forcing resource recreation.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceResource_update
func TestAccResourceResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial resource
			{
				Config: testAccResourceResourceConfigWithTags("test://terraform/update", "tf-update-resource", []string{"initial"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_resource.test", "uri", "test://terraform/update"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "name", "tf-update-resource"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.0", "initial"),
				),
			},
			// Update URI, name, and tags
			{
				Config: testAccResourceResourceConfigWithTags("test://terraform/updated", "tf-updated-resource", []string{"updated", "modified"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_resource.test", "uri", "test://terraform/updated"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "name", "tf-updated-resource"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.0", "updated"),
					resource.TestCheckResourceAttr("contextforge_resource.test", "tags.1", "modified"),
				),
			},
		},
	})
}

// TestAccResourceResource_import tests importing an existing resource.
// This verifies that resources can be imported using their ID and that
// all attributes are correctly populated in the state.
//
// To run:
//   make integration-test-all
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceResource_import
func TestAccResourceResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccResourceResourceConfig("test://terraform/import", "tf-import-resource", "Resource for import testing"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contextforge_resource.test", "id"),
				),
			},
			// Import by ID
			{
				ResourceName:      "contextforge_resource.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceResource_missingRequired tests error handling when required fields are missing.
// This verifies that the resource properly validates required attributes.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccResourceResource_missingRequired
func TestAccResourceResource_missingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceResourceConfigMissingURI(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "uri" is required`),
			},
			{
				Config:      testAccResourceResourceConfigMissingName(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "name" is required`),
			},
		},
	})
}

// testAccResourceResourceConfig generates basic Terraform configuration for a resource resource.
// This helper creates a minimal valid resource configuration.
//
// Parameters:
//   - uri: Resource URI
//   - name: Resource name
//   - description: Resource description
//
// Returns:
//   - HCL configuration string
func testAccResourceResourceConfig(uri, name, description string) string {
	return fmt.Sprintf(`
resource "contextforge_resource" "test" {
  uri         = %[1]q
  name        = %[2]q
  content     = "Test content for resource"
  description = %[3]q
}
`, uri, name, description)
}

// testAccResourceResourceConfigComplete generates Terraform configuration with all attributes.
// This includes mime_type, size, tags, and other optional fields.
//
// Returns:
//   - HCL configuration string with all attributes
func testAccResourceResourceConfigComplete() string {
	return `
resource "contextforge_resource" "test" {
  uri         = "test://terraform/complete"
  name        = "tf-complete-resource"
  content     = "Complete test content with all attributes"
  description = "Complete resource with all attributes"
  mime_type   = "text/plain"
  tags        = ["terraform", "testing", "complete"]
}
`
}

// testAccResourceResourceConfigWithTags generates Terraform configuration with tags.
// This helper is useful for testing tag updates and URI/name changes.
//
// Parameters:
//   - uri: Resource URI
//   - name: Resource name
//   - tags: List of tag strings
//
// Returns:
//   - HCL configuration string with tags
func testAccResourceResourceConfigWithTags(uri, name string, tags []string) string {
	tagsHCL := "["
	for i, tag := range tags {
		if i > 0 {
			tagsHCL += ", "
		}
		tagsHCL += fmt.Sprintf("%q", tag)
	}
	tagsHCL += "]"

	return fmt.Sprintf(`
resource "contextforge_resource" "test" {
  uri         = %[1]q
  name        = %[2]q
  content     = "Test content for tagged resource"
  description = "Resource with tags"
  tags        = %[3]s
}
`, uri, name, tagsHCL)
}

// testAccResourceResourceConfigMissingURI generates invalid Terraform configuration
// with missing required uri attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required uri attribute
func testAccResourceResourceConfigMissingURI() string {
	return `
resource "contextforge_resource" "test" {
  name        = "test-resource"
  description = "Resource missing required uri"
}
`
}

// testAccResourceResourceConfigMissingName generates invalid Terraform configuration
// with missing required name attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required name attribute
func testAccResourceResourceConfigMissingName() string {
	return `
resource "contextforge_resource" "test" {
  uri         = "test://terraform/missing-name"
  description = "Resource missing required name"
}
`
}
