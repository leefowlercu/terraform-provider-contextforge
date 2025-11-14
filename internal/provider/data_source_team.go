package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type teamDataSource struct {
	client *contextforge.Client
}

// Force compile-time validation that teamDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &teamDataSource{}

// Force compile-time validation that teamDataSource satisfies the datasource.DataSourceWithConfigure interface.
var _ datasource.DataSourceWithConfigure = &teamDataSource{}

// teamDataSourceModel defines the data source model.
type teamDataSourceModel struct {
	// Lookup field
	ID types.String `tfsdk:"id"`

	// Core fields
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	IsPersonal  types.Bool   `tfsdk:"is_personal"`
	Visibility  types.String `tfsdk:"visibility"`
	MaxMembers  types.Int64  `tfsdk:"max_members"`
	MemberCount types.Int64  `tfsdk:"member_count"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	CreatedBy   types.String `tfsdk:"created_by"`

	// Timestamps
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// NewTeamDataSource is a helper function to instantiate the team data source.
func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

// Metadata returns the data source type name.
func (d *teamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

// Schema defines the schema for the data source.
func (d *teamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge team by ID",
		Description:         "Data source for looking up a ContextForge team by ID",

		Attributes: map[string]schema.Attribute{
			// Lookup field
			"id": schema.StringAttribute{
				MarkdownDescription: "Team ID for lookup",
				Description:         "Team ID for lookup",
				Required:            true,
			},

			// Core fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Team name",
				Description:         "Team name",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Team slug (URL-friendly identifier)",
				Description:         "Team slug (URL-friendly identifier)",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Team description",
				Description:         "Team description",
				Computed:            true,
			},
			"is_personal": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a personal team",
				Description:         "Whether this is a personal team",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility setting (public, private, etc.)",
				Description:         "Visibility setting (public, private, etc.)",
				Computed:            true,
			},
			"max_members": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of members allowed in the team",
				Description:         "Maximum number of members allowed in the team",
				Computed:            true,
			},
			"member_count": schema.Int64Attribute{
				MarkdownDescription: "Current number of members in the team",
				Description:         "Current number of members in the team",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the team is active",
				Description:         "Whether the team is active",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Email address of the user who created the team",
				Description:         "Email address of the user who created the team",
				Computed:            true,
			},

			// Timestamps
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp (RFC3339 format)",
				Description:         "Creation timestamp (RFC3339 format)",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp (RFC3339 format)",
				Description:         "Last update timestamp (RFC3339 format)",
				Computed:            true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data teamDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ID is provided
	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up a team",
		)
		return
	}

	// Get team using List and filter (Get endpoint has authentication issues in v0.8.0)
	teams, _, err := d.client.Teams.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to List Teams", fmt.Sprintf("Unable to list teams; %v", err))
		return
	}

	var team *contextforge.Team
	targetID := data.ID.ValueString()
	for _, t := range teams {
		if t.ID == targetID {
			team = t
			break
		}
	}

	if team == nil {
		resp.Diagnostics.AddError("Team Not Found", fmt.Sprintf("Unable to find team with ID %s", targetID))
		return
	}

	// Map core fields
	data.ID = types.StringValue(team.ID)
	data.Name = types.StringValue(team.Name)
	data.Slug = types.StringValue(team.Slug)
	data.Description = types.StringPointerValue(team.Description)
	data.IsPersonal = types.BoolValue(team.IsPersonal)
	data.Visibility = types.StringPointerValue(team.Visibility)

	// Convert max_members (*int → types.Int64)
	if team.MaxMembers != nil {
		data.MaxMembers = types.Int64PointerValue(tfconv.Int64Ptr(*team.MaxMembers))
	} else {
		data.MaxMembers = types.Int64Null()
	}

	// Convert member_count (int → types.Int64)
	data.MemberCount = types.Int64Value(int64(team.MemberCount))

	data.IsActive = types.BoolValue(team.IsActive)
	data.CreatedBy = types.StringValue(team.CreatedBy)

	// Map timestamps (convert to RFC3339 string)
	if team.CreatedAt != nil && !team.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(team.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if team.UpdatedAt != nil && !team.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(team.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *teamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// Type assert the provider data to the expected client type
	client, ok := req.ProviderData.(*contextforge.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *contextforge.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Assign the client to the data source
	d.client = client
}
