package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccGetPromptID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-prompt-id.txt")
	if err != nil {
		t.Skipf("Prompt ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	promptID := strings.TrimSpace(string(data))
	if promptID == "" {
		t.Skip("Prompt ID file is empty - run 'make integration-test-setup' first")
	}

	return promptID
}

func TestAccPromptDataSource_basic(t *testing.T) {
	promptID := testAccGetPromptID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig(promptID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contextforge_prompt.test", "id", promptID),
					resource.TestCheckResourceAttr("data.contextforge_prompt.test", "name", "test-prompt"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "template"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "is_active"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccPromptDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

func TestAccPromptDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptDataSourceConfig("999999"),
				ExpectError: regexp.MustCompile(`Prompt Not Found|Unable to find prompt`),
			},
		},
	})
}

func TestAccPromptDataSource_allAttributes(t *testing.T) {
	promptID := testAccGetPromptID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig(promptID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contextforge_prompt.test", "id", promptID),
					resource.TestCheckResourceAttr("data.contextforge_prompt.test", "name", "test-prompt"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "template"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "is_active"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "updated_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "arguments.#"),
					resource.TestCheckResourceAttrSet("data.contextforge_prompt.test", "metrics.total_executions"),
				),
			},
		},
	})
}

func testAccPromptDataSourceConfig(promptID string) string {
	return fmt.Sprintf(`
data "contextforge_prompt" "test" {
  id = %[1]s
}
`, promptID)
}

func testAccPromptDataSourceConfigMissingID() string {
	return `
data "contextforge_prompt" "test" {
  # Missing required id attribute
}
`
}
