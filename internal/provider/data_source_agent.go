package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type agentDataSource struct {
	client *contextforge.Client
}

// Force compile-time validation that agentDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &agentDataSource{}

// Force compile-time validation that agentDataSource satisfies the datasource.DataSourceWithConfigure interface.
var _ datasource.DataSourceWithConfigure = &agentDataSource{}

// agentDataSourceModel defines the data source model.
type agentDataSourceModel struct {
	// Lookup field
	ID types.String `tfsdk:"id"`

	// Core fields
	Name            types.String  `tfsdk:"name"`
	Slug            types.String  `tfsdk:"slug"`
	Description     types.String  `tfsdk:"description"`
	EndpointURL     types.String  `tfsdk:"endpoint_url"`
	AgentType       types.String  `tfsdk:"agent_type"`
	ProtocolVersion types.String  `tfsdk:"protocol_version"`
	Capabilities    types.Dynamic `tfsdk:"capabilities"`
	Config          types.Dynamic `tfsdk:"config"`
	AuthType        types.String  `tfsdk:"auth_type"`
	Enabled         types.Bool    `tfsdk:"enabled"`
	Reachable       types.Bool    `tfsdk:"reachable"`

	// Nested metrics
	Metrics types.Object `tfsdk:"metrics"`

	// Organizational fields
	Tags       types.List   `tfsdk:"tags"`
	TeamID     types.String `tfsdk:"team_id"`
	OwnerEmail types.String `tfsdk:"owner_email"`
	Visibility types.String `tfsdk:"visibility"`

	// Timestamps
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	LastInteraction types.String `tfsdk:"last_interaction"`

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

// agentMetricsModel defines the nested metrics model.
type agentMetricsModel struct {
	TotalExecutions      types.Int64   `tfsdk:"total_executions"`
	SuccessfulExecutions types.Int64   `tfsdk:"successful_executions"`
	FailedExecutions     types.Int64   `tfsdk:"failed_executions"`
	FailureRate          types.Float64 `tfsdk:"failure_rate"`
	MinResponseTime      types.Float64 `tfsdk:"min_response_time"`
	MaxResponseTime      types.Float64 `tfsdk:"max_response_time"`
	AvgResponseTime      types.Float64 `tfsdk:"avg_response_time"`
	LastExecutionTime    types.String  `tfsdk:"last_execution_time"`
}

// NewAgentDataSource is a helper function to instantiate the agent data source.
func NewAgentDataSource() datasource.DataSource {
	return &agentDataSource{}
}

// Metadata returns the data source type name.
func (d *agentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

// Schema defines the schema for the data source.
func (d *agentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge A2A (Agent-to-Agent) agent by ID",
		Description:         "Data source for looking up a ContextForge A2A (Agent-to-Agent) agent by ID",

		Attributes: map[string]schema.Attribute{
			// Lookup field
			"id": schema.StringAttribute{
				MarkdownDescription: "Agent ID for lookup",
				Description:         "Agent ID for lookup",
				Required:            true,
			},

			// Core fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Agent name",
				Description:         "Agent name",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Agent slug (URL-friendly identifier)",
				Description:         "Agent slug (URL-friendly identifier)",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Agent description",
				Description:         "Agent description",
				Computed:            true,
			},
			"endpoint_url": schema.StringAttribute{
				MarkdownDescription: "Agent endpoint URL",
				Description:         "Agent endpoint URL",
				Computed:            true,
			},
			"agent_type": schema.StringAttribute{
				MarkdownDescription: "Agent type",
				Description:         "Agent type",
				Computed:            true,
			},
			"protocol_version": schema.StringAttribute{
				MarkdownDescription: "A2A protocol version",
				Description:         "A2A protocol version",
				Computed:            true,
			},
			"capabilities": schema.DynamicAttribute{
				MarkdownDescription: "Agent capabilities",
				Description:         "Agent capabilities",
				Computed:            true,
			},
			"config": schema.DynamicAttribute{
				MarkdownDescription: "Agent configuration",
				Description:         "Agent configuration",
				Computed:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type",
				Description:         "Authentication type",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent is enabled",
				Description:         "Whether the agent is enabled",
				Computed:            true,
			},
			"reachable": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent is reachable",
				Description:         "Whether the agent is reachable",
				Computed:            true,
			},

			// Nested metrics
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

			// Organizational fields
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Agent tags",
				Description:         "Agent tags",
				Computed:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Computed:            true,
			},
			"owner_email": schema.StringAttribute{
				MarkdownDescription: "Owner email address",
				Description:         "Owner email address",
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
			"last_interaction": schema.StringAttribute{
				MarkdownDescription: "Last interaction timestamp (RFC3339 format)",
				Description:         "Last interaction timestamp (RFC3339 format)",
				Computed:            true,
			},

			// Metadata (read-only)
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

// Read refreshes the Terraform state with the latest data.
func (d *agentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data agentDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ID is provided
	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up an agent",
		)
		return
	}

	// Get agent from API
	agent, _, err := d.client.Agents.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Agent",
			fmt.Sprintf("Unable to read agent with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

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
			resp.Diagnostics.AddError(
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
			resp.Diagnostics.AddError(
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
		metricsObject, diags := types.ObjectValueFrom(ctx, metricsModel.attrTypes(), metricsModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metrics = metricsObject
	} else {
		data.Metrics = types.ObjectNull(agentMetricsModel{}.attrTypes())
	}

	// Map organizational fields
	if agent.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, agent.Tags)
		resp.Diagnostics.Append(diags...)
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

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *agentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// attrTypes returns the attribute types map for agentMetricsModel.
func (m agentMetricsModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"total_executions":      types.Int64Type,
		"successful_executions": types.Int64Type,
		"failed_executions":     types.Int64Type,
		"failure_rate":          types.Float64Type,
		"min_response_time":     types.Float64Type,
		"max_response_time":     types.Float64Type,
		"avg_response_time":     types.Float64Type,
		"last_execution_time":   types.StringType,
	}
}
