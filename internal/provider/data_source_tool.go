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

type toolDataSource struct {
	client *contextforge.Client
}

// Force compile-time validation that toolDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &toolDataSource{}

// Force compile-time validation that toolDataSource satisfies the datasource.DataSourceWithConfigure interface.
var _ datasource.DataSourceWithConfigure = &toolDataSource{}

// toolDataSourceModel defines the data source model.
type toolDataSourceModel struct {
	// Lookup field
	ID types.String `tfsdk:"id"`

	// Core fields
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	InputSchema types.Dynamic `tfsdk:"input_schema"`
	Enabled     types.Bool    `tfsdk:"enabled"`

	// Organizational fields
	Tags       types.List   `tfsdk:"tags"`
	TeamID     types.String `tfsdk:"team_id"`
	Visibility types.String `tfsdk:"visibility"`

	// Timestamps
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// NewToolDataSource is a helper function to instantiate the tool data source.
func NewToolDataSource() datasource.DataSource {
	return &toolDataSource{}
}

// Metadata returns the data source type name.
func (d *toolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

// Schema defines the schema for the data source.
func (d *toolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge tool by ID",
		Description:         "Data source for looking up a ContextForge tool by ID",

		Attributes: map[string]schema.Attribute{
			// Lookup field
			"id": schema.StringAttribute{
				MarkdownDescription: "Tool ID for lookup",
				Description:         "Tool ID for lookup",
				Required:            true,
			},

			// Core fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Tool name",
				Description:         "Tool name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Tool description",
				Description:         "Tool description",
				Computed:            true,
			},
			"input_schema": schema.DynamicAttribute{
				MarkdownDescription: "JSON Schema defining tool input parameters",
				Description:         "JSON Schema defining tool input parameters",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the tool is enabled",
				Description:         "Whether the tool is enabled",
				Computed:            true,
			},

			// Organizational fields
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tool tags",
				Description:         "Tool tags",
				Computed:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility setting (public, private, etc.)",
				Description:         "Visibility setting (public, private, etc.)",
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
func (d *toolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data toolDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ID is provided
	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up a tool",
		)
		return
	}

	// Get tool from API
	tool, _, err := d.client.Tools.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Tool",
			fmt.Sprintf("Unable to read tool with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map core fields (note: tool.ID is string, not pointer)
	data.ID = types.StringValue(tool.ID)
	data.Name = types.StringValue(tool.Name)
	data.Description = types.StringPointerValue(tool.Description)
	data.Enabled = types.BoolValue(tool.Enabled)

	// Map input_schema (map[string]any -> types.Dynamic)
	if tool.InputSchema != nil {
		schemaValue, err := tfconv.ConvertMapToObjectValue(ctx, tool.InputSchema)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Input Schema",
				fmt.Sprintf("Unable to convert input_schema to object value; %v", err),
			)
			return
		}
		data.InputSchema = types.DynamicValue(schemaValue)
	} else {
		data.InputSchema = types.DynamicNull()
	}

	// Map organizational fields
	if tool.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(tool.Tags))
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(tool.TeamID)

	// IMPORTANT: Visibility is string, not *string (use StringValue not StringPointerValue)
	data.Visibility = types.StringValue(tool.Visibility)

	// Map timestamps (convert to RFC3339 string)
	if tool.CreatedAt != nil && !tool.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(tool.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if tool.UpdatedAt != nil && !tool.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(tool.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *toolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
