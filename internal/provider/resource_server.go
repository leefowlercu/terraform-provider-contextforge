package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
)

type serverResource struct {
	client *contextforge.Client
}

// Force compile-time validation that serverResource satisfies the resource.Resource interface.
var _ resource.Resource = &serverResource{}

// Force compile-time validation that serverResource satisfies the resource.ResourceWithConfigure interface.
var _ resource.ResourceWithConfigure = &serverResource{}

// Force compile-time validation that serverResource satisfies the resource.ResourceWithImportState interface.
var _ resource.ResourceWithImportState = &serverResource{}

// serverResourceModel defines the resource model.
type serverResourceModel struct {
	// Core fields
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Icon        types.String `tfsdk:"icon"`
	IsActive    types.Bool   `tfsdk:"is_active"`

	// Association fields
	AssociatedTools     types.List `tfsdk:"associated_tools"`
	AssociatedResources types.List `tfsdk:"associated_resources"`
	AssociatedPrompts   types.List `tfsdk:"associated_prompts"`
	AssociatedA2aAgents types.List `tfsdk:"associated_a2a_agents"`

	// Nested metrics
	Metrics types.Object `tfsdk:"metrics"`

	// Organizational fields
	Tags       types.List   `tfsdk:"tags"`
	TeamID     types.String `tfsdk:"team_id"`
	Team       types.String `tfsdk:"team"`
	OwnerEmail types.String `tfsdk:"owner_email"`
	Visibility types.String `tfsdk:"visibility"`

	// Timestamps
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`

	// Metadata (read-only)
	CreatedBy         types.String `tfsdk:"created_by"`
	CreatedFromIP     types.String `tfsdk:"created_from_ip"`
	CreatedVia        types.String `tfsdk:"created_via"`
	CreatedUserAgent  types.String `tfsdk:"created_user_agent"`
	ModifiedBy        types.String `tfsdk:"modified_by"`
	ModifiedFromIP    types.String `tfsdk:"modified_from_ip"`
	ModifiedVia       types.String `tfsdk:"modified_via"`
	ModifiedUserAgent types.String `tfsdk:"modified_user_agent"`
	ImportBatchID     types.String `tfsdk:"import_batch_id"`
	FederationSource  types.String `tfsdk:"federation_source"`
	Version           types.Int64  `tfsdk:"version"`
}

// NewServerResource is a helper function to instantiate the server resource.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

// Metadata returns the resource type name.
func (r *serverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

// Schema defines the schema for the resource.
func (r *serverResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge virtual server resource",
		Description:         "Manages a ContextForge virtual server resource",

		Attributes: map[string]schema.Attribute{
			// Core fields
			"id": schema.StringAttribute{
				MarkdownDescription: "Server unique identifier",
				Description:         "Server unique identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Description:         "Server name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Server description",
				Description:         "Server description",
				Optional:            true,
				Computed:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "Server icon (URL or data URI)",
				Description:         "Server icon (URL or data URI)",
				Optional:            true,
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the server is active",
				Description:         "Whether the server is active",
				Computed:            true,
			},

			// Association fields
			"associated_tools": schema.ListAttribute{
				MarkdownDescription: "List of associated tool IDs",
				Description:         "List of associated tool IDs",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"associated_resources": schema.ListAttribute{
				MarkdownDescription: "List of associated resource IDs",
				Description:         "List of associated resource IDs",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"associated_prompts": schema.ListAttribute{
				MarkdownDescription: "List of associated prompt IDs",
				Description:         "List of associated prompt IDs",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"associated_a2a_agents": schema.ListAttribute{
				MarkdownDescription: "List of associated A2A agent IDs",
				Description:         "List of associated A2A agent IDs",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},

			// Nested metrics
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "Server execution metrics",
				Description:         "Server execution metrics",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"total_executions": schema.Int64Attribute{
						MarkdownDescription: "Total number of executions",
						Description:         "Total number of executions",
						Computed:            true,
					},
					"successful_executions": schema.Int64Attribute{
						MarkdownDescription: "Number of successful executions",
						Description:         "Number of successful executions",
						Computed:            true,
					},
					"failed_executions": schema.Int64Attribute{
						MarkdownDescription: "Number of failed executions",
						Description:         "Number of failed executions",
						Computed:            true,
					},
					"failure_rate": schema.Float64Attribute{
						MarkdownDescription: "Failure rate (0-1)",
						Description:         "Failure rate (0-1)",
						Computed:            true,
					},
					"min_response_time": schema.Float64Attribute{
						MarkdownDescription: "Minimum response time in milliseconds",
						Description:         "Minimum response time in milliseconds",
						Computed:            true,
					},
					"max_response_time": schema.Float64Attribute{
						MarkdownDescription: "Maximum response time in milliseconds",
						Description:         "Maximum response time in milliseconds",
						Computed:            true,
					},
					"avg_response_time": schema.Float64Attribute{
						MarkdownDescription: "Average response time in milliseconds",
						Description:         "Average response time in milliseconds",
						Computed:            true,
					},
					"last_execution_time": schema.StringAttribute{
						MarkdownDescription: "Timestamp of last execution (RFC3339)",
						Description:         "Timestamp of last execution (RFC3339)",
						Computed:            true,
					},
				},
			},

			// Organizational fields
			"tags": schema.ListAttribute{
				MarkdownDescription: "Server tags",
				Description:         "Server tags",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Optional:            true,
				Computed:            true,
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "Team name (read-only)",
				Description:         "Team name (read-only)",
				Computed:            true,
			},
			"owner_email": schema.StringAttribute{
				MarkdownDescription: "Owner email address",
				Description:         "Owner email address",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Server visibility setting",
				Description:         "Server visibility setting",
				Optional:            true,
				Computed:            true,
			},

			// Timestamps
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp (RFC3339)",
				Description:         "Creation timestamp (RFC3339)",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp (RFC3339)",
				Description:         "Last update timestamp (RFC3339)",
				Computed:            true,
			},

			// Metadata (read-only)
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User who created the server",
				Description:         "User who created the server",
				Computed:            true,
			},
			"created_from_ip": schema.StringAttribute{
				MarkdownDescription: "IP address of creator",
				Description:         "IP address of creator",
				Computed:            true,
			},
			"created_via": schema.StringAttribute{
				MarkdownDescription: "Creation method",
				Description:         "Creation method",
				Computed:            true,
			},
			"created_user_agent": schema.StringAttribute{
				MarkdownDescription: "User agent of creator",
				Description:         "User agent of creator",
				Computed:            true,
			},
			"modified_by": schema.StringAttribute{
				MarkdownDescription: "User who last modified the server",
				Description:         "User who last modified the server",
				Computed:            true,
			},
			"modified_from_ip": schema.StringAttribute{
				MarkdownDescription: "IP address of modifier",
				Description:         "IP address of modifier",
				Computed:            true,
			},
			"modified_via": schema.StringAttribute{
				MarkdownDescription: "Modification method",
				Description:         "Modification method",
				Computed:            true,
			},
			"modified_user_agent": schema.StringAttribute{
				MarkdownDescription: "User agent of modifier",
				Description:         "User agent of modifier",
				Computed:            true,
			},
			"import_batch_id": schema.StringAttribute{
				MarkdownDescription: "Import batch identifier",
				Description:         "Import batch identifier",
				Computed:            true,
			},
			"federation_source": schema.StringAttribute{
				MarkdownDescription: "Federation source",
				Description:         "Federation source",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Server version number",
				Description:         "Server version number",
				Computed:            true,
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *serverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*contextforge.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *contextforge.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the server resource.
func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build ServerCreate from plan
	server := &contextforge.ServerCreate{
		Name: data.Name.ValueString(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		server.Description = &desc
	}

	if !data.Icon.IsNull() && !data.Icon.IsUnknown() {
		icon := data.Icon.ValueString()
		server.Icon = &icon
	}

	// Convert tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.Tags = tags
	}

	// Convert associated tools (string IDs)
	if !data.AssociatedTools.IsNull() && !data.AssociatedTools.IsUnknown() {
		var tools []string
		diags := data.AssociatedTools.ElementsAs(ctx, &tools, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedTools = tools
	}

	// Convert associated resources (string IDs)
	if !data.AssociatedResources.IsNull() && !data.AssociatedResources.IsUnknown() {
		var resources []string
		diags := data.AssociatedResources.ElementsAs(ctx, &resources, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedResources = resources
	}

	// Convert associated prompts (string IDs)
	if !data.AssociatedPrompts.IsNull() && !data.AssociatedPrompts.IsUnknown() {
		var prompts []string
		diags := data.AssociatedPrompts.ElementsAs(ctx, &prompts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedPrompts = prompts
	}

	// Convert associated A2A agents (string IDs)
	if !data.AssociatedA2aAgents.IsNull() && !data.AssociatedA2aAgents.IsUnknown() {
		var agents []string
		diags := data.AssociatedA2aAgents.ElementsAs(ctx, &agents, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedA2aAgents = agents
	}

	// Prepare create options for team_id and visibility
	opts := &contextforge.ServerCreateOptions{}
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		opts.TeamID = &teamID
	}
	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		opts.Visibility = &visibility
	}

	// Create server via API
	createdServer, _, err := r.client.Servers.Create(ctx, server, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Server",
			fmt.Sprintf("Unable to create server; %v", err),
		)
		return
	}

	// Map created server to state
	data.ID = types.StringValue(createdServer.ID)
	data.Name = types.StringValue(createdServer.Name)
	data.Description = types.StringPointerValue(createdServer.Description)
	data.Icon = types.StringPointerValue(createdServer.Icon)
	data.IsActive = types.BoolValue(createdServer.IsActive)

	// Map associations
	if createdServer.AssociatedTools != nil {
		toolsList, diags := types.ListValueFrom(ctx, types.StringType, createdServer.AssociatedTools)
		resp.Diagnostics.Append(diags...)
		data.AssociatedTools = toolsList
	} else {
		data.AssociatedTools = types.ListNull(types.StringType)
	}

	if createdServer.AssociatedResources != nil {
		resourcesList, diags := types.ListValueFrom(ctx, types.StringType, createdServer.AssociatedResources)
		resp.Diagnostics.Append(diags...)
		data.AssociatedResources = resourcesList
	} else {
		data.AssociatedResources = types.ListNull(types.StringType)
	}

	if createdServer.AssociatedPrompts != nil {
		promptsList, diags := types.ListValueFrom(ctx, types.StringType, createdServer.AssociatedPrompts)
		resp.Diagnostics.Append(diags...)
		data.AssociatedPrompts = promptsList
	} else {
		data.AssociatedPrompts = types.ListNull(types.StringType)
	}

	if createdServer.AssociatedA2aAgents != nil {
		agentsList, diags := types.ListValueFrom(ctx, types.StringType, createdServer.AssociatedA2aAgents)
		resp.Diagnostics.Append(diags...)
		data.AssociatedA2aAgents = agentsList
	} else {
		data.AssociatedA2aAgents = types.ListNull(types.StringType)
	}

	// Map metrics
	if createdServer.Metrics != nil {
		metricsObj, metricsDiags := mapMetricsToObject(ctx, createdServer.Metrics)
		resp.Diagnostics.Append(metricsDiags...)
		data.Metrics = metricsObj
	} else {
		attrTypes := map[string]attr.Type{
			"total_executions":      types.Int64Type,
			"successful_executions": types.Int64Type,
			"failed_executions":     types.Int64Type,
			"failure_rate":          types.Float64Type,
			"min_response_time":     types.Float64Type,
			"max_response_time":     types.Float64Type,
			"avg_response_time":     types.Float64Type,
			"last_execution_time":   types.StringType,
		}
		data.Metrics = types.ObjectNull(attrTypes)
	}

	// Map organizational fields
	if createdServer.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(createdServer.Tags))
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(createdServer.TeamID)
	data.Team = types.StringPointerValue(createdServer.Team)
	data.OwnerEmail = types.StringPointerValue(createdServer.OwnerEmail)
	data.Visibility = types.StringPointerValue(createdServer.Visibility)

	// Map timestamps
	if createdServer.CreatedAt != nil && !createdServer.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(createdServer.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if createdServer.UpdatedAt != nil && !createdServer.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(createdServer.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Map metadata
	data.CreatedBy = types.StringPointerValue(createdServer.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(createdServer.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(createdServer.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(createdServer.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(createdServer.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(createdServer.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(createdServer.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(createdServer.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(createdServer.ImportBatchID)
	data.FederationSource = types.StringPointerValue(createdServer.FederationSource)

	if createdServer.Version != nil {
		data.Version = types.Int64Value(int64(*createdServer.Version))
	} else {
		data.Version = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the server resource.
func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get server from API
	server, httpResp, err := r.client.Servers.Get(ctx, data.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read Server",
			fmt.Sprintf("Unable to read server; %v", err),
		)
		return
	}

	// Map server to state (same as Create)
	data.Name = types.StringValue(server.Name)
	data.Description = types.StringPointerValue(server.Description)
	data.Icon = types.StringPointerValue(server.Icon)
	data.IsActive = types.BoolValue(server.IsActive)

	// Map associations
	if server.AssociatedTools != nil {
		toolsList, diags := types.ListValueFrom(ctx, types.StringType, server.AssociatedTools)
		resp.Diagnostics.Append(diags...)
		data.AssociatedTools = toolsList
	} else {
		data.AssociatedTools = types.ListNull(types.StringType)
	}

	if server.AssociatedResources != nil {
		resourcesList, diags := types.ListValueFrom(ctx, types.StringType, server.AssociatedResources)
		resp.Diagnostics.Append(diags...)
		data.AssociatedResources = resourcesList
	} else {
		data.AssociatedResources = types.ListNull(types.StringType)
	}

	if server.AssociatedPrompts != nil {
		promptsList, diags := types.ListValueFrom(ctx, types.StringType, server.AssociatedPrompts)
		resp.Diagnostics.Append(diags...)
		data.AssociatedPrompts = promptsList
	} else {
		data.AssociatedPrompts = types.ListNull(types.StringType)
	}

	if server.AssociatedA2aAgents != nil {
		agentsList, diags := types.ListValueFrom(ctx, types.StringType, server.AssociatedA2aAgents)
		resp.Diagnostics.Append(diags...)
		data.AssociatedA2aAgents = agentsList
	} else {
		data.AssociatedA2aAgents = types.ListNull(types.StringType)
	}

	// Map metrics
	if server.Metrics != nil {
		metricsObj, metricsDiags := mapMetricsToObject(ctx, server.Metrics)
		resp.Diagnostics.Append(metricsDiags...)
		data.Metrics = metricsObj
	} else {
		attrTypes := map[string]attr.Type{
			"total_executions":      types.Int64Type,
			"successful_executions": types.Int64Type,
			"failed_executions":     types.Int64Type,
			"failure_rate":          types.Float64Type,
			"min_response_time":     types.Float64Type,
			"max_response_time":     types.Float64Type,
			"avg_response_time":     types.Float64Type,
			"last_execution_time":   types.StringType,
		}
		data.Metrics = types.ObjectNull(attrTypes)
	}

	// Map organizational fields
	if server.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(server.Tags))
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(server.TeamID)
	data.Team = types.StringPointerValue(server.Team)
	data.OwnerEmail = types.StringPointerValue(server.OwnerEmail)
	data.Visibility = types.StringPointerValue(server.Visibility)

	// Map timestamps
	if server.CreatedAt != nil && !server.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(server.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if server.UpdatedAt != nil && !server.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(server.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Map metadata
	data.CreatedBy = types.StringPointerValue(server.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(server.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(server.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(server.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(server.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(server.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(server.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(server.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(server.ImportBatchID)
	data.FederationSource = types.StringPointerValue(server.FederationSource)

	if server.Version != nil {
		data.Version = types.Int64Value(int64(*server.Version))
	} else {
		data.Version = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the server resource.
func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build ServerUpdate from plan
	server := &contextforge.ServerUpdate{}

	// Name update
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		server.Name = &name
	}

	// Description update
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		server.Description = &desc
	}

	// Icon update
	if !data.Icon.IsNull() && !data.Icon.IsUnknown() {
		icon := data.Icon.ValueString()
		server.Icon = &icon
	}

	// Tags update
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.Tags = tags
	}

	// Associated tools update (string IDs)
	if !data.AssociatedTools.IsNull() && !data.AssociatedTools.IsUnknown() {
		var tools []string
		diags := data.AssociatedTools.ElementsAs(ctx, &tools, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedTools = tools
	}

	// Associated resources update (string IDs)
	if !data.AssociatedResources.IsNull() && !data.AssociatedResources.IsUnknown() {
		var resources []string
		diags := data.AssociatedResources.ElementsAs(ctx, &resources, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedResources = resources
	}

	// Associated prompts update (string IDs)
	if !data.AssociatedPrompts.IsNull() && !data.AssociatedPrompts.IsUnknown() {
		var prompts []string
		diags := data.AssociatedPrompts.ElementsAs(ctx, &prompts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedPrompts = prompts
	}

	// Associated A2A agents update (string IDs)
	if !data.AssociatedA2aAgents.IsNull() && !data.AssociatedA2aAgents.IsUnknown() {
		var agents []string
		diags := data.AssociatedA2aAgents.ElementsAs(ctx, &agents, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.AssociatedA2aAgents = agents
	}

	// TeamID and Visibility updates
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		server.TeamID = &teamID
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		server.Visibility = &visibility
	}

	// Update server via API
	updatedServer, httpResp, err := r.client.Servers.Update(ctx, data.ID.ValueString(), server)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Update Server",
			fmt.Sprintf("Unable to update server; %v", err),
		)
		return
	}

	// Map updated server to state (same as Read)
	data.Name = types.StringValue(updatedServer.Name)
	data.Description = types.StringPointerValue(updatedServer.Description)
	data.Icon = types.StringPointerValue(updatedServer.Icon)
	data.IsActive = types.BoolValue(updatedServer.IsActive)

	// Map associations
	if updatedServer.AssociatedTools != nil {
		toolsList, diags := types.ListValueFrom(ctx, types.StringType, updatedServer.AssociatedTools)
		resp.Diagnostics.Append(diags...)
		data.AssociatedTools = toolsList
	} else {
		data.AssociatedTools = types.ListNull(types.StringType)
	}

	if updatedServer.AssociatedResources != nil {
		resourcesList, diags := types.ListValueFrom(ctx, types.StringType, updatedServer.AssociatedResources)
		resp.Diagnostics.Append(diags...)
		data.AssociatedResources = resourcesList
	} else {
		data.AssociatedResources = types.ListNull(types.StringType)
	}

	if updatedServer.AssociatedPrompts != nil {
		promptsList, diags := types.ListValueFrom(ctx, types.StringType, updatedServer.AssociatedPrompts)
		resp.Diagnostics.Append(diags...)
		data.AssociatedPrompts = promptsList
	} else {
		data.AssociatedPrompts = types.ListNull(types.StringType)
	}

	if updatedServer.AssociatedA2aAgents != nil {
		agentsList, diags := types.ListValueFrom(ctx, types.StringType, updatedServer.AssociatedA2aAgents)
		resp.Diagnostics.Append(diags...)
		data.AssociatedA2aAgents = agentsList
	} else {
		data.AssociatedA2aAgents = types.ListNull(types.StringType)
	}

	// Map metrics
	if updatedServer.Metrics != nil {
		metricsObj, metricsDiags := mapMetricsToObject(ctx, updatedServer.Metrics)
		resp.Diagnostics.Append(metricsDiags...)
		data.Metrics = metricsObj
	} else {
		attrTypes := map[string]attr.Type{
			"total_executions":      types.Int64Type,
			"successful_executions": types.Int64Type,
			"failed_executions":     types.Int64Type,
			"failure_rate":          types.Float64Type,
			"min_response_time":     types.Float64Type,
			"max_response_time":     types.Float64Type,
			"avg_response_time":     types.Float64Type,
			"last_execution_time":   types.StringType,
		}
		data.Metrics = types.ObjectNull(attrTypes)
	}

	// Map organizational fields
	if updatedServer.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(updatedServer.Tags))
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(updatedServer.TeamID)
	data.Team = types.StringPointerValue(updatedServer.Team)
	data.OwnerEmail = types.StringPointerValue(updatedServer.OwnerEmail)
	data.Visibility = types.StringPointerValue(updatedServer.Visibility)

	// Map timestamps
	if updatedServer.CreatedAt != nil && !updatedServer.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(updatedServer.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if updatedServer.UpdatedAt != nil && !updatedServer.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(updatedServer.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Map metadata
	data.CreatedBy = types.StringPointerValue(updatedServer.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(updatedServer.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(updatedServer.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(updatedServer.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(updatedServer.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(updatedServer.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(updatedServer.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(updatedServer.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(updatedServer.ImportBatchID)
	data.FederationSource = types.StringPointerValue(updatedServer.FederationSource)

	if updatedServer.Version != nil {
		data.Version = types.Int64Value(int64(*updatedServer.Version))
	} else {
		data.Version = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the server resource.
func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete server via API
	httpResp, err := r.client.Servers.Delete(ctx, data.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Already deleted, consider success
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Server",
			fmt.Sprintf("Unable to delete server; %v", err),
		)
		return
	}
}

// ImportState imports the server resource by ID.
func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID from the import request as the server ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions

// mapMetricsToObject converts ServerMetrics to types.Object.
func mapMetricsToObject(ctx context.Context, metrics *contextforge.ServerMetrics) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	metricsModel := serverMetricsModel{
		TotalExecutions:      types.Int64Value(int64(metrics.TotalExecutions)),
		SuccessfulExecutions: types.Int64Value(int64(metrics.SuccessfulExecutions)),
		FailedExecutions:     types.Int64Value(int64(metrics.FailedExecutions)),
		FailureRate:          types.Float64Value(metrics.FailureRate),
	}

	// Handle optional float64 pointers
	if metrics.MinResponseTime != nil {
		metricsModel.MinResponseTime = types.Float64Value(*metrics.MinResponseTime)
	} else {
		metricsModel.MinResponseTime = types.Float64Null()
	}

	if metrics.MaxResponseTime != nil {
		metricsModel.MaxResponseTime = types.Float64Value(*metrics.MaxResponseTime)
	} else {
		metricsModel.MaxResponseTime = types.Float64Null()
	}

	if metrics.AvgResponseTime != nil {
		metricsModel.AvgResponseTime = types.Float64Value(*metrics.AvgResponseTime)
	} else {
		metricsModel.AvgResponseTime = types.Float64Null()
	}

	// Handle timestamp
	if metrics.LastExecutionTime != nil && !metrics.LastExecutionTime.Time.IsZero() {
		metricsModel.LastExecutionTime = types.StringValue(metrics.LastExecutionTime.Time.Format(time.RFC3339))
	} else {
		metricsModel.LastExecutionTime = types.StringNull()
	}

	attrTypes := map[string]attr.Type{
		"total_executions":      types.Int64Type,
		"successful_executions": types.Int64Type,
		"failed_executions":     types.Int64Type,
		"failure_rate":          types.Float64Type,
		"min_response_time":     types.Float64Type,
		"max_response_time":     types.Float64Type,
		"avg_response_time":     types.Float64Type,
		"last_execution_time":   types.StringType,
	}

	objValue, objDiags := types.ObjectValueFrom(ctx, attrTypes, metricsModel)
	diags.Append(objDiags...)

	return objValue, diags
}
