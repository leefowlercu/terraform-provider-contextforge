package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type toolResource struct {
	client *contextforge.Client
}

// Force compile-time validation that toolResource satisfies the resource.Resource interface.
var _ resource.Resource = &toolResource{}

// Force compile-time validation that toolResource satisfies the resource.ResourceWithConfigure interface.
var _ resource.ResourceWithConfigure = &toolResource{}

// Force compile-time validation that toolResource satisfies the resource.ResourceWithImportState interface.
var _ resource.ResourceWithImportState = &toolResource{}

// toolResourceModel defines the resource model.
type toolResourceModel struct {
	// Computed field
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

// NewToolResource is a helper function to instantiate the tool resource.
func NewToolResource() resource.Resource {
	return &toolResource{}
}

// isEmptyInputSchema checks if the input schema is the default empty schema returned by the API.
// The API returns {"type": "object", "properties": {}} when no schema is set.
func isEmptyInputSchema(schema map[string]any) bool {
	if len(schema) == 0 {
		return true
	}
	if len(schema) == 2 {
		typeVal, hasType := schema["type"]
		propsVal, hasProps := schema["properties"]
		if hasType && hasProps && typeVal == "object" {
			if propsMap, ok := propsVal.(map[string]any); ok && len(propsMap) == 0 {
				return true
			}
		}
	}
	return false
}

// Metadata returns the resource type name.
func (r *toolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

// Schema defines the schema for the resource.
func (r *toolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge tool resource",
		Description:         "Manages a ContextForge tool resource",

		Attributes: map[string]schema.Attribute{
			// Computed field
			"id": schema.StringAttribute{
				MarkdownDescription: "Tool ID",
				Description:         "Tool ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Core fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Tool name",
				Description:         "Tool name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Tool description",
				Description:         "Tool description",
				Optional:            true,
			},
			"input_schema": schema.DynamicAttribute{
				MarkdownDescription: "JSON Schema defining tool input parameters",
				Description:         "JSON Schema defining tool input parameters",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the tool is enabled (default: true)",
				Description:         "Whether the tool is enabled (default: true)",
				Optional:            true,
			},

			// Organizational fields
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tool tags",
				Description:         "Tool tags",
				Optional:            true,
				Computed:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Optional:            true,
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility setting (public, private, etc.)",
				Description:         "Visibility setting (public, private, etc.)",
				Optional:            true,
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

// Create creates the resource and sets the initial Terraform state.
func (r *toolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data toolResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build tool object for API
	tool := &contextforge.Tool{
		Name:    data.Name.ValueString(),
		Enabled: data.Enabled.ValueBoolPointer() != nil && *data.Enabled.ValueBoolPointer(),
	}

	// Handle default for enabled if not set
	if data.Enabled.IsNull() || data.Enabled.IsUnknown() {
		tool.Enabled = true
	}

	// Map optional description
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		tool.Description = &desc
	}

	// Map optional input_schema (types.Dynamic -> map[string]any)
	if !data.InputSchema.IsNull() && !data.InputSchema.IsUnknown() {
		schemaMap, err := tfconv.ConvertObjectValueToMap(ctx, data.InputSchema.UnderlyingValue())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Input Schema",
				fmt.Sprintf("Unable to convert input_schema from object value; %v", err),
			)
			return
		}
		tool.InputSchema = schemaMap
	}

	// Map optional tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tool.Tags = tags
	}

	// Prepare create options for team_id and visibility
	opts := &contextforge.ToolCreateOptions{}

	// Map optional team_id to create options
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		opts.TeamID = &teamID
	}

	// Map optional visibility to create options
	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		opts.Visibility = &visibility
	}

	// Create tool via API
	createdTool, _, err := r.client.Tools.Create(ctx, tool, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Tool",
			fmt.Sprintf("Unable to create tool; %v", err),
		)
		return
	}

	// Map response to state
	data.ID = types.StringValue(createdTool.ID)
	data.Name = types.StringValue(createdTool.Name)
	data.Description = types.StringPointerValue(createdTool.Description)
	data.Enabled = types.BoolValue(createdTool.Enabled)

	// Map input_schema from response
	// Only set input_schema if it was explicitly provided in the config or if it's non-empty
	if createdTool.InputSchema != nil && !isEmptyInputSchema(createdTool.InputSchema) {
		schemaValue, err := tfconv.ConvertMapToObjectValue(ctx, createdTool.InputSchema)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Input Schema",
				fmt.Sprintf("Unable to convert input_schema to object value; %v", err),
			)
			return
		}
		data.InputSchema = types.DynamicValue(schemaValue)
	} else if !data.InputSchema.IsNull() && !data.InputSchema.IsUnknown() {
		// Preserve plan value if it was set
		// data.InputSchema already has plan value
	} else {
		data.InputSchema = types.DynamicNull()
	}

	// Map tags from response
	if createdTool.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, createdTool.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(createdTool.TeamID)
	data.Visibility = types.StringValue(createdTool.Visibility)

	// Map timestamps
	if createdTool.CreatedAt != nil && !createdTool.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(createdTool.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if createdTool.UpdatedAt != nil && !createdTool.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(createdTool.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *toolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data toolResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get tool from API
	tool, httpResp, err := r.client.Tools.Get(ctx, data.ID.ValueString())
	if err != nil {
		// Handle 404 - resource no longer exists
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read Tool",
			fmt.Sprintf("Unable to read tool with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map response to state
	data.ID = types.StringValue(tool.ID)
	data.Name = types.StringValue(tool.Name)
	data.Description = types.StringPointerValue(tool.Description)
	data.Enabled = types.BoolValue(tool.Enabled)

	// Map input_schema
	// Only set input_schema if it's not the default empty schema
	if tool.InputSchema != nil && !isEmptyInputSchema(tool.InputSchema) {
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

	// Map tags
	if tool.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, tool.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(tool.TeamID)
	data.Visibility = types.StringValue(tool.Visibility)

	// Map timestamps
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

// Update updates the resource and sets the updated Terraform state on success.
func (r *toolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data toolResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build tool object for API
	tool := &contextforge.Tool{
		Name:    data.Name.ValueString(),
		Enabled: data.Enabled.ValueBoolPointer() != nil && *data.Enabled.ValueBoolPointer(),
	}

	// Handle default for enabled if not set
	if data.Enabled.IsNull() || data.Enabled.IsUnknown() {
		tool.Enabled = true
	}

	// Map optional description
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		tool.Description = &desc
	}

	// Map optional input_schema
	if !data.InputSchema.IsNull() && !data.InputSchema.IsUnknown() {
		schemaMap, err := tfconv.ConvertObjectValueToMap(ctx, data.InputSchema.UnderlyingValue())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Input Schema",
				fmt.Sprintf("Unable to convert input_schema from object value; %v", err),
			)
			return
		}
		tool.InputSchema = schemaMap
	}

	// Map optional tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tool.Tags = tags
	}

	// Map optional team_id (Update uses tool struct directly)
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		tool.TeamID = &teamID
	}

	// Map optional visibility (Update uses tool struct directly)
	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		tool.Visibility = data.Visibility.ValueString()
	}

	// Update tool via API
	_, httpResp, err := r.client.Tools.Update(ctx, data.ID.ValueString(), tool)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Tool",
			fmt.Sprintf("Unable to update tool with ID %s (status: %d); %v", data.ID.ValueString(), httpResp.StatusCode, err),
		)
		return
	}

	// Read the updated tool (API returns old resource from Update endpoint)
	updatedTool, _, err := r.client.Tools.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Updated Tool",
			fmt.Sprintf("Unable to read tool after update (ID: %s); %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map response to state
	data.ID = types.StringValue(updatedTool.ID)
	data.Name = types.StringValue(updatedTool.Name)
	data.Description = types.StringPointerValue(updatedTool.Description)
	data.Enabled = types.BoolValue(updatedTool.Enabled)

	// Map input_schema from response
	// Only set input_schema if it's not the default empty schema
	if updatedTool.InputSchema != nil && !isEmptyInputSchema(updatedTool.InputSchema) {
		schemaValue, err := tfconv.ConvertMapToObjectValue(ctx, updatedTool.InputSchema)
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

	// Map tags from response
	if updatedTool.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, updatedTool.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(updatedTool.TeamID)
	data.Visibility = types.StringValue(updatedTool.Visibility)

	// Map timestamps
	if updatedTool.CreatedAt != nil && !updatedTool.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(updatedTool.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if updatedTool.UpdatedAt != nil && !updatedTool.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(updatedTool.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *toolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data toolResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete tool via API
	httpResp, err := r.client.Tools.Delete(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore 404 errors (resource already deleted)
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Tool",
			fmt.Sprintf("Unable to delete tool with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// State is automatically removed by the framework
}

// Configure adds the provider configured client to the resource.
func (r *toolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// Type assert the provider data to the expected client type
	client, ok := req.ProviderData.(*contextforge.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *contextforge.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Assign the client to the resource
	r.client = client
}

// ImportState imports an existing resource by ID.
func (r *toolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID from the import request as the tool ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
