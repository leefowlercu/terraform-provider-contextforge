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

type resourceDataSource struct {
	client *contextforge.Client
}

// Force compile-time validation that resourceDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &resourceDataSource{}

// Force compile-time validation that resourceDataSource satisfies the datasource.DataSourceWithConfigure interface.
var _ datasource.DataSourceWithConfigure = &resourceDataSource{}

// resourceDataSourceModel defines the data source model.
type resourceDataSourceModel struct {
	// Lookup field
	ID types.String `tfsdk:"id"`

	// Core fields
	URI         types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	MimeType    types.String `tfsdk:"mime_type"`
	Size        types.Int64  `tfsdk:"size"`
	IsActive    types.Bool   `tfsdk:"is_active"`

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

// resourceMetricsModel defines the nested metrics model.
type resourceMetricsModel struct {
	TotalExecutions      types.Int64   `tfsdk:"total_executions"`
	SuccessfulExecutions types.Int64   `tfsdk:"successful_executions"`
	FailedExecutions     types.Int64   `tfsdk:"failed_executions"`
	FailureRate          types.Float64 `tfsdk:"failure_rate"`
	MinResponseTime      types.Float64 `tfsdk:"min_response_time"`
	MaxResponseTime      types.Float64 `tfsdk:"max_response_time"`
	AvgResponseTime      types.Float64 `tfsdk:"avg_response_time"`
	LastExecutionTime    types.String  `tfsdk:"last_execution_time"`
}

// NewResourceDataSource is a helper function to instantiate the resource data source.
func NewResourceDataSource() datasource.DataSource {
	return &resourceDataSource{}
}

// Metadata returns the data source type name.
func (d *resourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

// Schema defines the schema for the data source.
func (d *resourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge resource by ID",
		Description:         "Data source for looking up a ContextForge resource by ID",

		Attributes: map[string]schema.Attribute{
			// Lookup field
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource ID for lookup",
				Description:         "Resource ID for lookup",
				Required:            true,
			},

			// Core fields
			"uri": schema.StringAttribute{
				MarkdownDescription: "Resource URI",
				Description:         "Resource URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name",
				Description:         "Resource name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Resource description",
				Description:         "Resource description",
				Computed:            true,
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "Resource MIME type",
				Description:         "Resource MIME type",
				Computed:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Resource size in bytes",
				Description:         "Resource size in bytes",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the resource is active",
				Description:         "Whether the resource is active",
				Computed:            true,
			},

			// Nested metrics
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "Resource performance metrics",
				Description:         "Resource performance metrics",
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
				MarkdownDescription: "Resource tags",
				Description:         "Resource tags",
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
				MarkdownDescription: "User who created the resource (read-only metadata)",
				Description:         "User who created the resource (read-only metadata)",
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
				MarkdownDescription: "User who last modified the resource (read-only metadata)",
				Description:         "User who last modified the resource (read-only metadata)",
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
func (d *resourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resourceDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ID is provided
	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up a resource",
		)
		return
	}

	// Get resource from API using List and filter
	// Note: The API doesn't have a dedicated metadata endpoint for resources by ID,
	// so we must use List() and filter by ID
	resources, _, err := d.client.Resources.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to List Resources",
			fmt.Sprintf("Unable to list resources; %v", err),
		)
		return
	}

	// Find the resource by ID
	var resource *contextforge.Resource
	targetID := data.ID.ValueString()
	for _, r := range resources {
		if r.ID != nil && r.ID.String() == targetID {
			resource = r
			break
		}
	}

	if resource == nil {
		resp.Diagnostics.AddError(
			"Resource Not Found",
			fmt.Sprintf("Unable to find resource with ID %s", targetID),
		)
		return
	}

	// Map core fields
	// Note: resource.ID is *FlexibleID, need to use .String() method
	if resource.ID != nil {
		data.ID = types.StringValue(resource.ID.String())
	}
	data.URI = types.StringValue(resource.URI)
	data.Name = types.StringValue(resource.Name)
	data.Description = types.StringPointerValue(resource.Description)
	data.MimeType = types.StringPointerValue(resource.MimeType)

	// Convert *int to types.Int64
	if resource.Size != nil {
		data.Size = types.Int64PointerValue(tfconv.Int64Ptr(*resource.Size))
	} else {
		data.Size = types.Int64Null()
	}

	data.IsActive = types.BoolValue(resource.IsActive)

	// Map nested metrics
	if resource.Metrics != nil {
		metricsModel := resourceMetricsModel{
			TotalExecutions:      types.Int64Value(int64(resource.Metrics.TotalExecutions)),
			SuccessfulExecutions: types.Int64Value(int64(resource.Metrics.SuccessfulExecutions)),
			FailedExecutions:     types.Int64Value(int64(resource.Metrics.FailedExecutions)),
			FailureRate:          types.Float64Value(resource.Metrics.FailureRate),
		}

		// Handle optional float64 pointers
		if resource.Metrics.MinResponseTime != nil {
			metricsModel.MinResponseTime = types.Float64Value(*resource.Metrics.MinResponseTime)
		} else {
			metricsModel.MinResponseTime = types.Float64Null()
		}

		if resource.Metrics.MaxResponseTime != nil {
			metricsModel.MaxResponseTime = types.Float64Value(*resource.Metrics.MaxResponseTime)
		} else {
			metricsModel.MaxResponseTime = types.Float64Null()
		}

		if resource.Metrics.AvgResponseTime != nil {
			metricsModel.AvgResponseTime = types.Float64Value(*resource.Metrics.AvgResponseTime)
		} else {
			metricsModel.AvgResponseTime = types.Float64Null()
		}

		// Handle optional timestamp
		if resource.Metrics.LastExecutionTime != nil && !resource.Metrics.LastExecutionTime.Time.IsZero() {
			metricsModel.LastExecutionTime = types.StringValue(resource.Metrics.LastExecutionTime.Time.Format(time.RFC3339))
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
		data.Metrics = types.ObjectNull(resourceMetricsModel{}.attrTypes())
	}

	// Map organizational fields
	if resource.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, resource.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(resource.TeamID)
	data.Team = types.StringPointerValue(resource.Team)
	data.OwnerEmail = types.StringPointerValue(resource.OwnerEmail)
	data.Visibility = types.StringPointerValue(resource.Visibility)

	// Map timestamps (convert to RFC3339 string)
	if resource.CreatedAt != nil && !resource.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(resource.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if resource.UpdatedAt != nil && !resource.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(resource.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Map metadata fields
	data.CreatedBy = types.StringPointerValue(resource.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(resource.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(resource.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(resource.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(resource.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(resource.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(resource.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(resource.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(resource.ImportBatchID)
	data.FederationSource = types.StringPointerValue(resource.FederationSource)

	if resource.Version != nil {
		data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*resource.Version))
	} else {
		data.Version = types.Int64Null()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *resourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// attrTypes returns the attribute types map for resourceMetricsModel.
func (m resourceMetricsModel) attrTypes() map[string]attr.Type {
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
