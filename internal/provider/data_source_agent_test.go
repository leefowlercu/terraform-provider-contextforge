package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccGetAgentID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-agent-id.txt")
	if err != nil {
		t.Skipf("Agent ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	agentID := strings.TrimSpace(string(data))
	if agentID == "" {
		t.Skip("Agent ID file is empty - run 'make integration-test-setup' first")
	}

	return agentID
}

func TestAccAgentDataSource_basic(t *testing.T) {
	agentID := testAccGetAgentID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentDataSourceConfig(agentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contextforge_agent.test", "id", agentID),
					resource.TestCheckResourceAttr("data.contextforge_agent.test", "name", "test-agent"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "endpoint_url"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccAgentDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgentDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

func TestAccAgentDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgentDataSourceConfig("non-existent-agent-id-12345"),
				ExpectError: regexp.MustCompile(`Failed to Read Agent|Unable to read agent`),
			},
		},
	})
}

func TestAccAgentDataSource_allAttributes(t *testing.T) {
	agentID := testAccGetAgentID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentDataSourceConfig(agentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contextforge_agent.test", "id", agentID),
					resource.TestCheckResourceAttr("data.contextforge_agent.test", "name", "test-agent"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "tags.#"),
					resource.TestCheckResourceAttrSet("data.contextforge_agent.test", "metrics.total_executions"),
				),
			},
		},
	})
}

func testAccAgentDataSourceConfig(agentID string) string {
	return fmt.Sprintf(`
data "contextforge_agent" "test" {
  id = %[1]q
}
`, agentID)
}

func testAccAgentDataSourceConfigMissingID() string {
	return `
data "contextforge_agent" "test" {
  # Missing required id attribute
}
`
}
