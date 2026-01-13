package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type agentResource struct {
	client *contextforge.Client
}

// Force compile-time validation that agentResource satisfies the resource.Resource interface.
var _ resource.Resource = &agentResource{}

// Force compile-time validation that agentResource satisfies the resource.ResourceWithImportState interface.
var _ resource.ResourceWithImportState = &agentResource{}

// Force compile-time validation that agentResource satisfies the resource.ResourceWithConfigure interface.
var _ resource.ResourceWithConfigure = &agentResource{}

// agentResourceModel defines the resource model.
type agentResourceModel struct {
	// Computed fields
	ID   types.String `tfsdk:"id"`
	Slug types.String `tfsdk:"slug"`

	// Required fields
	Name        types.String `tfsdk:"name"`
	EndpointURL types.String `tfsdk:"endpoint_url"`

	// Optional fields
	Description     types.String  `tfsdk:"description"`
	AgentType       types.String  `tfsdk:"agent_type"`
	ProtocolVersion types.String  `tfsdk:"protocol_version"`
	Config          types.Dynamic `tfsdk:"config"`
	AuthType        types.String  `tfsdk:"auth_type"`
	Enabled         types.Bool    `tfsdk:"enabled"`
	Tags            types.List    `tfsdk:"tags"`
	TeamID          types.String  `tfsdk:"team_id"`
	Visibility      types.String  `tfsdk:"visibility"`

	// Computed fields
	Capabilities types.Dynamic `tfsdk:"capabilities"`
	Reachable    types.Bool    `tfsdk:"reachable"`
	OwnerEmail   types.String  `tfsdk:"owner_email"`

	// Nested metrics (computed)
	Metrics types.Object `tfsdk:"metrics"`

	// Timestamps (computed)
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	LastInteraction types.String `tfsdk:"last_interaction"`

	// Metadata (computed, read-only)
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

// NewAgentResource is a helper function to instantiate the agent resource.
func NewAgentResource() resource.Resource {
	return &agentResource{}
}

// Metadata returns the resource type name.
func (r *agentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

// Schema defines the schema for the resource.
func (r *agentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge A2A (Agent-to-Agent) agent resource",
		Description:         "Manages a ContextForge A2A (Agent-to-Agent) agent resource",

		Attributes: map[string]schema.Attribute{
			// Computed ID
			"id": schema.StringAttribute{
				MarkdownDescription: "Agent ID",
				Description:         "Agent ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Agent slug (URL-friendly identifier)",
				Description:         "Agent slug (URL-friendly identifier)",
				Computed:            true,
			},

			// Required fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Agent name",
				Description:         "Agent name",
				Required:            true,
			},
			"endpoint_url": schema.StringAttribute{
				MarkdownDescription: "Agent endpoint URL",
				Description:         "Agent endpoint URL",
				Required:            true,
			},

			// Optional fields
			"description": schema.StringAttribute{
				MarkdownDescription: "Agent description",
				Description:         "Agent description",
				Optional:            true,
			},
			"agent_type": schema.StringAttribute{
				MarkdownDescription: "Agent type",
				Description:         "Agent type",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"protocol_version": schema.StringAttribute{
				MarkdownDescription: "A2A protocol version",
				Description:         "A2A protocol version",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"config": schema.DynamicAttribute{
				MarkdownDescription: "Agent configuration (arbitrary JSON object)",
				Description:         "Agent configuration (arbitrary JSON object)",
				Optional:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type",
				Description:         "Authentication type",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent is enabled (defaults to true)",
				Description:         "Whether the agent is enabled (defaults to true)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Agent tags",
				Description:         "Agent tags",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility setting (public, private, etc.)",
				Description:         "Visibility setting (public, private, etc.)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Computed fields
			"capabilities": schema.DynamicAttribute{
				MarkdownDescription: "Agent capabilities (computed from agent metadata)",
				Description:         "Agent capabilities (computed from agent metadata)",
				Computed:            true,
			},
			"reachable": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent is reachable",
				Description:         "Whether the agent is reachable",
				Computed:            true,
			},
			"owner_email": schema.StringAttribute{
				MarkdownDescription: "Owner email address",
				Description:         "Owner email address",
				Computed:            true,
			},

			// Nested metrics (computed)
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "Agent performance metrics",
				Description:         "Agent performance metrics",
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
						MarkdownDescription: "Failure rate (0.0 to 1.0)",
						Description:         "Failure rate (0.0 to 1.0)",
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
						MarkdownDescription: "Last execution timestamp (RFC3339 format)",
						Description:         "Last execution timestamp (RFC3339 format)",
						Computed:            true,
					},
				},
			},

			// Timestamps (computed)
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
			"last_interaction": schema.StringAttribute{
				MarkdownDescription: "Last interaction timestamp (RFC3339 format)",
				Description:         "Last interaction timestamp (RFC3339 format)",
				Computed:            true,
			},

			// Metadata (computed, read-only)
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User who created the agent (read-only metadata)",
				Description:         "User who created the agent (read-only metadata)",
				Computed:            true,
			},
			"created_from_ip": schema.StringAttribute{
				MarkdownDescription: "IP address of creator (read-only metadata)",
				Description:         "IP address of creator (read-only metadata)",
				Computed:            true,
			},
			"created_via": schema.StringAttribute{
				MarkdownDescription: "Creation method (read-only metadata)",
				Description:         "Creation method (read-only metadata)",
				Computed:            true,
			},
			"created_user_agent": schema.StringAttribute{
				MarkdownDescription: "User agent of creator (read-only metadata)",
				Description:         "User agent of creator (read-only metadata)",
				Computed:            true,
			},
			"modified_by": schema.StringAttribute{
				MarkdownDescription: "User who last modified the agent (read-only metadata)",
				Description:         "User who last modified the agent (read-only metadata)",
				Computed:            true,
			},
			"modified_from_ip": schema.StringAttribute{
				MarkdownDescription: "IP address of last modifier (read-only metadata)",
				Description:         "IP address of last modifier (read-only metadata)",
				Computed:            true,
			},
			"modified_via": schema.StringAttribute{
				MarkdownDescription: "Modification method (read-only metadata)",
				Description:         "Modification method (read-only metadata)",
				Computed:            true,
			},
			"modified_user_agent": schema.StringAttribute{
				MarkdownDescription: "User agent of last modifier (read-only metadata)",
				Description:         "User agent of last modifier (read-only metadata)",
				Computed:            true,
			},
			"import_batch_id": schema.StringAttribute{
				MarkdownDescription: "Import batch identifier (read-only metadata)",
				Description:         "Import batch identifier (read-only metadata)",
				Computed:            true,
			},
			"federation_source": schema.StringAttribute{
				MarkdownDescription: "Federation source (read-only metadata)",
				Description:         "Federation source (read-only metadata)",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number (read-only metadata)",
				Description:         "Version number (read-only metadata)",
				Computed:            true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *agentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data agentResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build agent create struct with required fields
	agent := &contextforge.AgentCreate{
		Name:        data.Name.ValueString(),
		EndpointURL: data.EndpointURL.ValueString(),
	}

	// Add optional fields with nil checks
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		agent.Description = &desc
	}

	if !data.AgentType.IsNull() && !data.AgentType.IsUnknown() {
		agentType := data.AgentType.ValueString()
		agent.AgentType = agentType
	}

	if !data.ProtocolVersion.IsNull() && !data.ProtocolVersion.IsUnknown() {
		protocolVersion := data.ProtocolVersion.ValueString()
		agent.ProtocolVersion = protocolVersion
	}

	// Convert config Dynamic to map[string]any
	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		configMap, err := tfconv.ConvertObjectValueToMap(ctx, data.Config.UnderlyingValue())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Config",
				fmt.Sprintf("Unable to convert config from object value; %v", err),
			)
			return
		}
		agent.Config = configMap
	}

	if !data.AuthType.IsNull() && !data.AuthType.IsUnknown() {
		authType := data.AuthType.ValueString()
		agent.AuthType = &authType
	}

	// Tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		agent.Tags = tags
	}

	// Build CreateOptions for team_id and visibility
	opts := &contextforge.AgentCreateOptions{}
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		opts.TeamID = &teamID
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		opts.Visibility = &visibility
	}

	// Create agent via API
	createdAgent, _, err := r.client.Agents.Create(ctx, agent, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Agent",
			fmt.Sprintf("Unable to create agent; %v", err),
		)
		return
	}

	// Map response to state using helper
	r.mapAgentToState(ctx, createdAgent, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *agentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data agentResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get agent from API
	agent, httpResp, err := r.client.Agents.Get(ctx, data.ID.ValueString())
	if err != nil {
		// Handle 404 as resource deleted
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read Agent",
			fmt.Sprintf("Unable to read agent with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map response to state using helper
	r.mapAgentToState(ctx, agent, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *agentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data agentResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build agent update struct (use three-state update semantics)
	agent := &contextforge.AgentUpdate{}

	// Required fields become optional in update
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		agent.Name = &name
	}

	if !data.EndpointURL.IsNull() && !data.EndpointURL.IsUnknown() {
		endpointURL := data.EndpointURL.ValueString()
		agent.EndpointURL = &endpointURL
	}

	// Optional fields
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		agent.Description = &desc
	}

	if !data.AgentType.IsNull() && !data.AgentType.IsUnknown() {
		agentType := data.AgentType.ValueString()
		agent.AgentType = &agentType
	}

	if !data.ProtocolVersion.IsNull() && !data.ProtocolVersion.IsUnknown() {
		protocolVersion := data.ProtocolVersion.ValueString()
		agent.ProtocolVersion = &protocolVersion
	}

	// Convert config Dynamic to map[string]any
	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		configMap, err := tfconv.ConvertObjectValueToMap(ctx, data.Config.UnderlyingValue())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Config",
				fmt.Sprintf("Unable to convert config from object value; %v", err),
			)
			return
		}
		agent.Config = configMap
	}

	if !data.AuthType.IsNull() && !data.AuthType.IsUnknown() {
		authType := data.AuthType.ValueString()
		agent.AuthType = &authType
	}

	// Tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		agent.Tags = tags
	}

	// Team/Visibility (Update uses agent struct directly, not options)
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		agent.TeamID = &teamID
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		agent.Visibility = &visibility
	}

	// Update agent via API
	updatedAgent, httpResp, err := r.client.Agents.Update(ctx, data.ID.ValueString(), agent)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Update Agent",
			fmt.Sprintf("Unable to update agent; %v", err),
		)
		return
	}

	// Map response to state using helper
	r.mapAgentToState(ctx, updatedAgent, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *agentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data agentResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete agent via API
	httpResp, err := r.client.Agents.Delete(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore 404 errors (resource already deleted)
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Agent",
			fmt.Sprintf("Unable to delete agent with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// State is automatically removed by the framework
}

// ImportState imports the resource state by ID.
func (r *agentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure adds the provider configured client to the resource.
func (r *agentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// mapAgentToState is a helper to map Agent API response to Terraform state.
// This helper handles all type conversions including Dynamic types and nested metrics.
func (r *agentResource) mapAgentToState(ctx context.Context, agent *contextforge.Agent, data *agentResourceModel, diags *diag.Diagnostics) {
	// Map core fields
	data.ID = types.StringValue(agent.ID)
	data.Name = types.StringValue(agent.Name)
	data.Slug = types.StringValue(agent.Slug)
	data.Description = types.StringPointerValue(agent.Description)
	data.EndpointURL = types.StringValue(agent.EndpointURL)
	data.AgentType = types.StringValue(agent.AgentType)
	data.ProtocolVersion = types.StringValue(agent.ProtocolVersion)

	// Map capabilities (map[string]any -> types.Dynamic)
	if agent.Capabilities != nil {
		capabilitiesValue, err := tfconv.ConvertMapToObjectValue(ctx, agent.Capabilities)
		if err != nil {
			diags.AddError(
				"Failed to Convert Capabilities",
				fmt.Sprintf("Unable to convert capabilities to object value; %v", err),
			)
			return
		}
		data.Capabilities = types.DynamicValue(capabilitiesValue)
	} else {
		data.Capabilities = types.DynamicNull()
	}

	// Map config (map[string]any -> types.Dynamic)
	if agent.Config != nil {
		configValue, err := tfconv.ConvertMapToObjectValue(ctx, agent.Config)
		if err != nil {
			diags.AddError(
				"Failed to Convert Config",
				fmt.Sprintf("Unable to convert config to object value; %v", err),
			)
			return
		}
		data.Config = types.DynamicValue(configValue)
	} else {
		data.Config = types.DynamicNull()
	}

	data.AuthType = types.StringPointerValue(agent.AuthType)
	data.Enabled = types.BoolValue(agent.Enabled)
	data.Reachable = types.BoolValue(agent.Reachable)

	// Map nested metrics
	if agent.Metrics != nil {
		metricsModel := agentMetricsModel{
			TotalExecutions:      types.Int64Value(int64(agent.Metrics.TotalExecutions)),
			SuccessfulExecutions: types.Int64Value(int64(agent.Metrics.SuccessfulExecutions)),
			FailedExecutions:     types.Int64Value(int64(agent.Metrics.FailedExecutions)),
			FailureRate:          types.Float64Value(agent.Metrics.FailureRate),
		}

		// Handle optional float64 pointers
		if agent.Metrics.MinResponseTime != nil {
			metricsModel.MinResponseTime = types.Float64Value(*agent.Metrics.MinResponseTime)
		} else {
			metricsModel.MinResponseTime = types.Float64Null()
		}

		if agent.Metrics.MaxResponseTime != nil {
			metricsModel.MaxResponseTime = types.Float64Value(*agent.Metrics.MaxResponseTime)
		} else {
			metricsModel.MaxResponseTime = types.Float64Null()
		}

		if agent.Metrics.AvgResponseTime != nil {
			metricsModel.AvgResponseTime = types.Float64Value(*agent.Metrics.AvgResponseTime)
		} else {
			metricsModel.AvgResponseTime = types.Float64Null()
		}

		// Handle optional timestamp
		if agent.Metrics.LastExecutionTime != nil && !agent.Metrics.LastExecutionTime.Time.IsZero() {
			metricsModel.LastExecutionTime = types.StringValue(agent.Metrics.LastExecutionTime.Time.Format(time.RFC3339))
		} else {
			metricsModel.LastExecutionTime = types.StringNull()
		}

		// Convert metrics model to object
		metricsObject, metricsDiags := types.ObjectValueFrom(ctx, metricsModel.attrTypes(), metricsModel)
		diags.Append(metricsDiags...)
		if diags.HasError() {
			return
		}
		data.Metrics = metricsObject
	} else {
		data.Metrics = types.ObjectNull(agentMetricsModel{}.attrTypes())
	}

	// Map organizational fields
	if agent.Tags != nil {
		tagsList, tagsDiags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(agent.Tags))
		diags.Append(tagsDiags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(agent.TeamID)
	data.OwnerEmail = types.StringPointerValue(agent.OwnerEmail)
	data.Visibility = types.StringPointerValue(agent.Visibility)

	// Map timestamps (convert to RFC3339 string)
	if agent.CreatedAt != nil && !agent.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(agent.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if agent.UpdatedAt != nil && !agent.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(agent.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	if agent.LastInteraction != nil && !agent.LastInteraction.Time.IsZero() {
		data.LastInteraction = types.StringValue(agent.LastInteraction.Time.Format(time.RFC3339))
	} else {
		data.LastInteraction = types.StringNull()
	}

	// Map metadata fields
	data.CreatedBy = types.StringPointerValue(agent.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(agent.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(agent.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(agent.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(agent.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(agent.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(agent.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(agent.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(agent.ImportBatchID)
	data.FederationSource = types.StringPointerValue(agent.FederationSource)

	if agent.Version != nil {
		data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*agent.Version))
	} else {
		data.Version = types.Int64Null()
	}
}
