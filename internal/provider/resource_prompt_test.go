package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("tf-prompt-basic", "Hello {{name}}", "Prompt created by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contextforge_prompt.test", "id"),
					resource.TestCheckResourceAttr("contextforge_prompt.test", "name", "tf-prompt-basic"),
					resource.TestCheckResourceAttr("contextforge_prompt.test", "template", "Hello {{name}}"),
					resource.TestCheckResourceAttr("contextforge_prompt.test", "description", "Prompt created by Terraform"),
					resource.TestCheckResourceAttrSet("contextforge_prompt.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_prompt.test", "updated_at"),
				),
			},
			{
				ResourceName:      "contextforge_prompt.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPromptResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("tf-prompt-update", "Hello {{name}}", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_prompt.test", "name", "tf-prompt-update"),
					resource.TestCheckResourceAttr("contextforge_prompt.test", "description", "Initial description"),
				),
			},
			{
				Config: testAccPromptResourceConfigWithCustomName("tf-prompt-update", "Hello {{name}}", "Renamed by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_prompt.test", "custom_name", "tf-prompt-custom"),
					resource.TestCheckResourceAttr("contextforge_prompt.test", "display_name", "Terraform Prompt"),
					resource.TestCheckResourceAttr("contextforge_prompt.test", "description", "Renamed by Terraform"),
				),
			},
		},
	})
}

func TestAccPromptResource_missingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptResourceConfigMissingName(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "name" is required`),
			},
			{
				Config:      testAccPromptResourceConfigMissingTemplate(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "template" is required`),
			},
		},
	})
}

func testAccPromptResourceConfig(name, template, description string) string {
	return fmt.Sprintf(`
resource "contextforge_prompt" "test" {
  name        = %[1]q
  template    = %[2]q
  description = %[3]q
  arguments = [
    {
      name        = "name"
      description = "Name to greet"
      required    = true
    }
  ]
}
`, name, template, description)
}

func testAccPromptResourceConfigWithCustomName(name, template, description string) string {
	return fmt.Sprintf(`
resource "contextforge_prompt" "test" {
  name         = %[1]q
  template     = %[2]q
  description  = %[3]q
  custom_name  = "tf-prompt-custom"
  display_name = "Terraform Prompt"
  arguments = [
    {
      name        = "name"
      description = "Name to greet"
      required    = true
    }
  ]
  tags = ["terraform", "prompt"]
}
`, name, template, description)
}

func testAccPromptResourceConfigMissingName() string {
	return `
resource "contextforge_prompt" "test" {
  template = "Hello {{name}}"
}
`
}

func testAccPromptResourceConfigMissingTemplate() string {
	return `
resource "contextforge_prompt" "test" {
  name = "missing-template"
}
`
}
