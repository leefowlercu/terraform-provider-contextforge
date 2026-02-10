package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamResource_basic(t *testing.T) {
	t.Skip("Skipping due to upstream team endpoint auth behavior: /teams/{id} and DELETE /teams/{id} require user auth and reject bearer tokens")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamResourceConfig("tf-team-basic", "Team created by Terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contextforge_team.test", "id"),
					resource.TestCheckResourceAttr("contextforge_team.test", "name", "tf-team-basic"),
					resource.TestCheckResourceAttr("contextforge_team.test", "description", "Team created by Terraform"),
					resource.TestCheckResourceAttrSet("contextforge_team.test", "slug"),
					resource.TestCheckResourceAttrSet("contextforge_team.test", "created_at"),
					resource.TestCheckResourceAttrSet("contextforge_team.test", "updated_at"),
				),
			},
			{
				ResourceName:      "contextforge_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTeamResource_update(t *testing.T) {
	t.Skip("Skipping due to upstream team endpoint auth behavior: /teams/{id} and PUT /teams/{id} require user auth and reject bearer tokens")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamResourceConfig("tf-team-update", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_team.test", "name", "tf-team-update"),
					resource.TestCheckResourceAttr("contextforge_team.test", "description", "Initial description"),
				),
			},
			{
				Config: testAccTeamResourceConfig("tf-team-updated", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contextforge_team.test", "name", "tf-team-updated"),
					resource.TestCheckResourceAttr("contextforge_team.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccTeamResource_missingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTeamResourceConfigMissingName(),
				ExpectError: regexp.MustCompile(`Missing required argument|The argument "name" is required`),
			},
		},
	})
}

func testAccTeamResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "contextforge_team" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccTeamResourceConfigMissingName() string {
	return `
resource "contextforge_team" "test" {
  description = "Team without name"
}
`
}
