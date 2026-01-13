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

type promptDataSource struct {
	client *contextforge.Client
}

var _ datasource.DataSource = &promptDataSource{}
var _ datasource.DataSourceWithConfigure = &promptDataSource{}

type promptDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Template    types.String `tfsdk:"template"`
	Arguments   types.List   `tfsdk:"arguments"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	Tags        types.List   `tfsdk:"tags"`
	Metrics     types.Object `tfsdk:"metrics"`
	TeamID      types.String `tfsdk:"team_id"`
	Team        types.String `tfsdk:"team"`
	OwnerEmail  types.String `tfsdk:"owner_email"`
	Visibility  types.String `tfsdk:"visibility"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`

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

type promptArgumentModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Required    types.Bool   `tfsdk:"required"`
}

type promptMetricsModel struct {
	TotalExecutions      types.Int64   `tfsdk:"total_executions"`
	SuccessfulExecutions types.Int64   `tfsdk:"successful_executions"`
	FailedExecutions     types.Int64   `tfsdk:"failed_executions"`
	FailureRate          types.Float64 `tfsdk:"failure_rate"`
	MinResponseTime      types.Float64 `tfsdk:"min_response_time"`
	MaxResponseTime      types.Float64 `tfsdk:"max_response_time"`
	AvgResponseTime      types.Float64 `tfsdk:"avg_response_time"`
	LastExecutionTime    types.String  `tfsdk:"last_execution_time"`
}

func NewPromptDataSource() datasource.DataSource {
	return &promptDataSource{}
}

func (d *promptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (d *promptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge prompt by ID",
		Description:         "Data source for looking up a ContextForge prompt by ID",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Prompt ID for lookup",
				Description:         "Prompt ID for lookup",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Prompt name",
				Description:         "Prompt name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Prompt description",
				Description:         "Prompt description",
				Computed:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Prompt template",
				Description:         "Prompt template",
				Computed:            true,
			},
			"arguments": schema.ListNestedAttribute{
				MarkdownDescription: "Prompt arguments/parameters",
				Description:         "Prompt arguments/parameters",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Argument name",
							Description:         "Argument name",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Argument description",
							Description:         "Argument description",
							Computed:            true,
						},
						"required": schema.BoolAttribute{
							MarkdownDescription: "Whether the argument is required",
							Description:         "Whether the argument is required",
							Computed:            true,
						},
					},
				},
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the prompt is active",
				Description:         "Whether the prompt is active",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Prompt tags",
				Description:         "Prompt tags",
				Computed:            true,
			},
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "Prompt performance metrics",
				Description:         "Prompt performance metrics",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"total_executions":      schema.Int64Attribute{Computed: true},
					"successful_executions": schema.Int64Attribute{Computed: true},
					"failed_executions":     schema.Int64Attribute{Computed: true},
					"failure_rate":          schema.Float64Attribute{Computed: true},
					"min_response_time":     schema.Float64Attribute{Computed: true},
					"max_response_time":     schema.Float64Attribute{Computed: true},
					"avg_response_time":     schema.Float64Attribute{Computed: true},
					"last_execution_time":   schema.StringAttribute{Computed: true},
				},
			},
			"team_id":     schema.StringAttribute{Computed: true},
			"team":        schema.StringAttribute{Computed: true},
			"owner_email": schema.StringAttribute{Computed: true},
			"visibility":  schema.StringAttribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},

			"created_by":          schema.StringAttribute{Computed: true},
			"created_from_ip":     schema.StringAttribute{Computed: true},
			"created_via":         schema.StringAttribute{Computed: true},
			"created_user_agent":  schema.StringAttribute{Computed: true},
			"modified_by":         schema.StringAttribute{Computed: true},
			"modified_from_ip":    schema.StringAttribute{Computed: true},
			"modified_via":        schema.StringAttribute{Computed: true},
			"modified_user_agent": schema.StringAttribute{Computed: true},
			"import_batch_id":     schema.StringAttribute{Computed: true},
			"federation_source":   schema.StringAttribute{Computed: true},
			"version":             schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *promptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data promptDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() {
		resp.Diagnostics.AddError("Missing Required Attribute", "The 'id' attribute must be specified")
		return
	}

	// Get prompt using List and filter (no Get metadata method)
	prompts, _, err := d.client.Prompts.List(ctx, &contextforge.PromptListOptions{IncludeInactive: true})
	if err != nil {
		resp.Diagnostics.AddError("Failed to List Prompts", fmt.Sprintf("Unable to list prompts; %v", err))
		return
	}

	var prompt *contextforge.Prompt
	targetID := data.ID.ValueString()
	for _, p := range prompts {
		if p.ID == targetID {
			prompt = p
			break
		}
	}

	if prompt == nil {
		resp.Diagnostics.AddError("Prompt Not Found", fmt.Sprintf("Unable to find prompt with ID %s", targetID))
		return
	}

	data.ID = types.StringValue(prompt.ID)
	data.Name = types.StringValue(prompt.Name)
	data.Description = types.StringPointerValue(prompt.Description)
	data.Template = types.StringValue(prompt.Template)
	data.IsActive = types.BoolValue(prompt.IsActive)

	// Map arguments
	if len(prompt.Arguments) > 0 {
		argModels := make([]promptArgumentModel, len(prompt.Arguments))
		for i, arg := range prompt.Arguments {
			argModels[i] = promptArgumentModel{
				Name:        types.StringValue(arg.Name),
				Description: types.StringPointerValue(arg.Description),
				Required:    types.BoolValue(arg.Required),
			}
		}
		argsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: promptArgumentModel{}.attrTypes()}, argModels)
		resp.Diagnostics.Append(diags...)
		data.Arguments = argsList
	} else {
		data.Arguments = types.ListNull(types.ObjectType{AttrTypes: promptArgumentModel{}.attrTypes()})
	}

	// Map metrics
	if prompt.Metrics != nil {
		metricsModel := promptMetricsModel{
			TotalExecutions:      types.Int64Value(int64(prompt.Metrics.TotalExecutions)),
			SuccessfulExecutions: types.Int64Value(int64(prompt.Metrics.SuccessfulExecutions)),
			FailedExecutions:     types.Int64Value(int64(prompt.Metrics.FailedExecutions)),
			FailureRate:          types.Float64Value(prompt.Metrics.FailureRate),
		}
		if prompt.Metrics.MinResponseTime != nil {
			metricsModel.MinResponseTime = types.Float64Value(*prompt.Metrics.MinResponseTime)
		} else {
			metricsModel.MinResponseTime = types.Float64Null()
		}
		if prompt.Metrics.MaxResponseTime != nil {
			metricsModel.MaxResponseTime = types.Float64Value(*prompt.Metrics.MaxResponseTime)
		} else {
			metricsModel.MaxResponseTime = types.Float64Null()
		}
		if prompt.Metrics.AvgResponseTime != nil {
			metricsModel.AvgResponseTime = types.Float64Value(*prompt.Metrics.AvgResponseTime)
		} else {
			metricsModel.AvgResponseTime = types.Float64Null()
		}
		if prompt.Metrics.LastExecutionTime != nil && !prompt.Metrics.LastExecutionTime.Time.IsZero() {
			metricsModel.LastExecutionTime = types.StringValue(prompt.Metrics.LastExecutionTime.Time.Format(time.RFC3339))
		} else {
			metricsModel.LastExecutionTime = types.StringNull()
		}
		metricsObject, diags := types.ObjectValueFrom(ctx, metricsModel.attrTypes(), metricsModel)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Metrics = metricsObject
		}
	} else {
		data.Metrics = types.ObjectNull(promptMetricsModel{}.attrTypes())
	}

	if prompt.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(prompt.Tags))
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(prompt.TeamID)
	data.Team = types.StringPointerValue(prompt.Team)
	data.OwnerEmail = types.StringPointerValue(prompt.OwnerEmail)
	data.Visibility = types.StringPointerValue(prompt.Visibility)

	if prompt.CreatedAt != nil && !prompt.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(prompt.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}
	if prompt.UpdatedAt != nil && !prompt.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(prompt.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	data.CreatedBy = types.StringPointerValue(prompt.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(prompt.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(prompt.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(prompt.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(prompt.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(prompt.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(prompt.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(prompt.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(prompt.ImportBatchID)
	data.FederationSource = types.StringPointerValue(prompt.FederationSource)

	if prompt.Version != nil {
		data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*prompt.Version))
	} else {
		data.Version = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *promptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*contextforge.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *contextforge.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (m promptArgumentModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"required":    types.BoolType,
	}
}

func (m promptMetricsModel) attrTypes() map[string]attr.Type {
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
