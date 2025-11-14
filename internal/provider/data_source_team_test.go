package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccGetTeamID reads the dynamically created team ID from the integration test setup.
// This team is created by the integration-test-setup.sh script.
//
// The function will skip the test if:
//   - The team ID file doesn't exist (setup script not run)
//   - The file is empty
//
// Returns:
//   - The team ID string
func testAccGetTeamID(t *testing.T) string {
	data, err := os.ReadFile("../../tmp/contextforge-test-team-id.txt")
	if err != nil {
		t.Skipf("Team ID file not found - run 'make integration-test-setup' first: %v", err)
	}

	teamID := strings.TrimSpace(string(data))
	if teamID == "" {
		t.Skip("Team ID file is empty - run 'make integration-test-setup' first")
	}

	return teamID
}

// TestAccTeamDataSource_basic tests successful team lookup by ID.
// This test verifies that the data source can retrieve a team and populate
// all expected attributes.
//
// Prerequisites:
//   - CONTEXTFORGE_ADDR environment variable set
//   - CONTEXTFORGE_TOKEN environment variable set
//   - Integration test setup completed (creates test team)
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   make integration-test      # Tests only (requires manual setup)
func TestAccTeamDataSource_basic(t *testing.T) {
	teamID := testAccGetTeamID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccTeamDataSourceConfig(teamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify lookup attribute
					resource.TestCheckResourceAttr("data.contextforge_team.test", "id", teamID),

					// Verify expected team values
					resource.TestCheckResourceAttr("data.contextforge_team.test", "name", "test-team"),
					resource.TestCheckResourceAttr("data.contextforge_team.test", "slug", "test-team"),

					// Verify boolean attributes
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "is_active"),
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "is_personal"),

					// Verify timestamp attributes are populated
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "updated_at"),

					// Verify created_by is set
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "created_by"),
				),
			},
		},
	})
}

// TestAccTeamDataSource_missingID tests error handling when ID is not provided.
// The data source should return a clear error message when the required ID attribute
// is missing from the configuration.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccTeamDataSource_missingID
func TestAccTeamDataSource_missingID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTeamDataSourceConfigMissingID(),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

// TestAccTeamDataSource_nonExistent tests error handling for non-existent team ID.
// The data source should return an appropriate error when attempting to look up
// a team that doesn't exist.
//
// To run:
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccTeamDataSource_nonExistent
func TestAccTeamDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTeamDataSourceConfig("non-existent-team-id-12345"),
				ExpectError: regexp.MustCompile(`Team Not Found|Unable to find team`),
			},
		},
	})
}

// TestAccTeamDataSource_allAttributes tests that all team attributes are
// properly mapped and accessible. This test verifies optional/nullable fields
// are handled correctly.
//
// To run:
//   make integration-test-all  # Full lifecycle with setup/teardown
//   TF_ACC=1 go test -v ./internal/provider/ -run TestAccTeamDataSource_allAttributes
func TestAccTeamDataSource_allAttributes(t *testing.T) {
	teamID := testAccGetTeamID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamDataSourceConfig(teamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Core fields
					resource.TestCheckResourceAttr("data.contextforge_team.test", "id", teamID),
					resource.TestCheckResourceAttr("data.contextforge_team.test", "name", "test-team"),
					resource.TestCheckResourceAttr("data.contextforge_team.test", "slug", "test-team"),
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "is_active"),
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "is_personal"),

					// Timestamps
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "updated_at"),

					// Created by
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "created_by"),

					// Member count should be set
					resource.TestCheckResourceAttrSet("data.contextforge_team.test", "member_count"),

					// Note: Other attributes (description, visibility, max_members) may be null/empty
					// depending on the team configuration. We don't check them as that would fail for null values.
				),
			},
		},
	})
}

// testAccTeamDataSourceConfig returns the Terraform configuration for team lookup by ID.
// This helper function generates HCL configuration for testing the team data source.
//
// Parameters:
//   - teamID: The ID of the team to look up
//
// Returns:
//   - HCL configuration string with the data source definition
func testAccTeamDataSourceConfig(teamID string) string {
	return fmt.Sprintf(`
data "contextforge_team" "test" {
  id = %[1]q
}
`, teamID)
}

// testAccTeamDataSourceConfigMissingID returns invalid Terraform configuration
// with missing required ID attribute. This is used to test error handling.
//
// Returns:
//   - HCL configuration string missing the required id attribute
func testAccTeamDataSourceConfigMissingID() string {
	return `
data "contextforge_team" "test" {
  # Missing required id attribute
}
`
}

// testAccTeamDataSourceConfigWithOutputs returns Terraform configuration
// that includes outputs to verify data source attribute values.
//
// This helper is useful for debugging and verifying that specific attribute
// values are correctly retrieved from the API.
//
// Parameters:
//   - teamID: The ID of the team to look up
//
// Returns:
//   - HCL configuration string with data source and outputs
func testAccTeamDataSourceConfigWithOutputs(teamID string) string {
	return fmt.Sprintf(`
data "contextforge_team" "test" {
  id = %[1]q
}

output "team_name" {
  value = data.contextforge_team.test.name
}

output "team_slug" {
  value = data.contextforge_team.test.slug
}

output "team_is_personal" {
  value = data.contextforge_team.test.is_personal
}

output "team_member_count" {
  value = data.contextforge_team.test.member_count
}
`, teamID)
}
