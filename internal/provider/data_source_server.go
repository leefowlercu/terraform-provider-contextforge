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

type serverDataSource struct {
	client *contextforge.Client
}

// Force compile-time validation that serverDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &serverDataSource{}

// Force compile-time validation that serverDataSource satisfies the datasource.DataSourceWithConfigure interface.
var _ datasource.DataSourceWithConfigure = &serverDataSource{}

// serverDataSourceModel defines the data source model.
type serverDataSourceModel struct {
	// Lookup field
	ID types.String `tfsdk:"id"`

	// Core fields
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

// serverMetricsModel defines the nested metrics model.
type serverMetricsModel struct {
	TotalExecutions      types.Int64   `tfsdk:"total_executions"`
	SuccessfulExecutions types.Int64   `tfsdk:"successful_executions"`
	FailedExecutions     types.Int64   `tfsdk:"failed_executions"`
	FailureRate          types.Float64 `tfsdk:"failure_rate"`
	MinResponseTime      types.Float64 `tfsdk:"min_response_time"`
	MaxResponseTime      types.Float64 `tfsdk:"max_response_time"`
	AvgResponseTime      types.Float64 `tfsdk:"avg_response_time"`
	LastExecutionTime    types.String  `tfsdk:"last_execution_time"`
}

// NewServerDataSource is a helper function to instantiate the server data source.
func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// Metadata returns the data source type name.
func (d *serverDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

// Schema defines the schema for the data source.
func (d *serverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge server by ID",
		Description:         "Data source for looking up a ContextForge server by ID",

		Attributes: map[string]schema.Attribute{
			// Lookup field
			"id": schema.StringAttribute{
				MarkdownDescription: "Server ID for lookup",
				Description:         "Server ID for lookup",
				Required:            true,
			},

			// Core fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Description:         "Server name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Server description",
				Description:         "Server description",
				Computed:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "Server icon (URL or data URI)",
				Description:         "Server icon (URL or data URI)",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the server is active",
				Description:         "Whether the server is active",
				Computed:            true,
			},

			// Association fields
			"associated_tools": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Associated tool IDs",
				Description:         "Associated tool IDs",
				Computed:            true,
			},
			"associated_resources": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Associated resource IDs",
				Description:         "Associated resource IDs",
				Computed:            true,
			},
			"associated_prompts": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Associated prompt IDs",
				Description:         "Associated prompt IDs",
				Computed:            true,
			},
			"associated_a2a_agents": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Associated A2A agent IDs",
				Description:         "Associated A2A agent IDs",
				Computed:            true,
			},

			// Nested metrics
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "Server performance metrics",
				Description:         "Server performance metrics",
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
				MarkdownDescription: "Server tags",
				Description:         "Server tags",
				Computed:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Computed:            true,
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "Team name",
				Description:         "Team name",
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

			// Metadata (read-only)
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User who created the server (read-only metadata)",
				Description:         "User who created the server (read-only metadata)",
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
				MarkdownDescription: "User who last modified the server (read-only metadata)",
				Description:         "User who last modified the server (read-only metadata)",
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
func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ID is provided
	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up a server",
		)
		return
	}

	// Get server from API
	server, _, err := d.client.Servers.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Server",
			fmt.Sprintf("Unable to read server with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map core fields (note: server.ID is string, not pointer)
	data.ID = types.StringValue(server.ID)
	data.Name = types.StringValue(server.Name)
	data.Description = types.StringPointerValue(server.Description)
	data.Icon = types.StringPointerValue(server.Icon)
	data.IsActive = types.BoolValue(server.IsActive)

	// Map association fields - string slices
	if server.AssociatedTools != nil {
		toolsList, diags := types.ListValueFrom(ctx, types.StringType, server.AssociatedTools)
		resp.Diagnostics.Append(diags...)
		data.AssociatedTools = toolsList
	} else {
		data.AssociatedTools = types.ListNull(types.StringType)
	}

	if server.AssociatedA2aAgents != nil {
		agentsList, diags := types.ListValueFrom(ctx, types.StringType, server.AssociatedA2aAgents)
		resp.Diagnostics.Append(diags...)
		data.AssociatedA2aAgents = agentsList
	} else {
		data.AssociatedA2aAgents = types.ListNull(types.StringType)
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

	// Map nested metrics
	if server.Metrics != nil {
		metricsModel := serverMetricsModel{
			TotalExecutions:      types.Int64Value(int64(server.Metrics.TotalExecutions)),
			SuccessfulExecutions: types.Int64Value(int64(server.Metrics.SuccessfulExecutions)),
			FailedExecutions:     types.Int64Value(int64(server.Metrics.FailedExecutions)),
			FailureRate:          types.Float64Value(server.Metrics.FailureRate),
		}

		// Handle optional float64 pointers
		if server.Metrics.MinResponseTime != nil {
			metricsModel.MinResponseTime = types.Float64Value(*server.Metrics.MinResponseTime)
		} else {
			metricsModel.MinResponseTime = types.Float64Null()
		}

		if server.Metrics.MaxResponseTime != nil {
			metricsModel.MaxResponseTime = types.Float64Value(*server.Metrics.MaxResponseTime)
		} else {
			metricsModel.MaxResponseTime = types.Float64Null()
		}

		if server.Metrics.AvgResponseTime != nil {
			metricsModel.AvgResponseTime = types.Float64Value(*server.Metrics.AvgResponseTime)
		} else {
			metricsModel.AvgResponseTime = types.Float64Null()
		}

		// Handle optional timestamp
		if server.Metrics.LastExecutionTime != nil && !server.Metrics.LastExecutionTime.Time.IsZero() {
			metricsModel.LastExecutionTime = types.StringValue(server.Metrics.LastExecutionTime.Time.Format(time.RFC3339))
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
		data.Metrics = types.ObjectNull(serverMetricsModel{}.attrTypes())
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

	// Map timestamps (convert to RFC3339 string)
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

	// Map metadata fields
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
		data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*server.Version))
	} else {
		data.Version = types.Int64Null()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *serverDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// attrTypes returns the attribute types map for serverMetricsModel.
func (m serverMetricsModel) attrTypes() map[string]attr.Type {
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
