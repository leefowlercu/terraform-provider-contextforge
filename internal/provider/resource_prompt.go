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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type promptResource struct {
	client *contextforge.Client
}

var _ resource.Resource = &promptResource{}
var _ resource.ResourceWithConfigure = &promptResource{}
var _ resource.ResourceWithImportState = &promptResource{}

type promptResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OriginalName   types.String `tfsdk:"original_name"`
	CustomName     types.String `tfsdk:"custom_name"`
	CustomNameSlug types.String `tfsdk:"custom_name_slug"`
	DisplayName    types.String `tfsdk:"display_name"`
	GatewaySlug    types.String `tfsdk:"gateway_slug"`
	Description    types.String `tfsdk:"description"`
	Template       types.String `tfsdk:"template"`
	Arguments      types.List   `tfsdk:"arguments"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Tags           types.List   `tfsdk:"tags"`
	Metrics        types.Object `tfsdk:"metrics"`
	TeamID         types.String `tfsdk:"team_id"`
	Team           types.String `tfsdk:"team"`
	OwnerEmail     types.String `tfsdk:"owner_email"`
	Visibility     types.String `tfsdk:"visibility"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`

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

type promptResourceArgumentModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Required    types.Bool   `tfsdk:"required"`
}

type promptResourceMetricsModel struct {
	TotalExecutions      types.Int64   `tfsdk:"total_executions"`
	SuccessfulExecutions types.Int64   `tfsdk:"successful_executions"`
	FailedExecutions     types.Int64   `tfsdk:"failed_executions"`
	FailureRate          types.Float64 `tfsdk:"failure_rate"`
	MinResponseTime      types.Float64 `tfsdk:"min_response_time"`
	MaxResponseTime      types.Float64 `tfsdk:"max_response_time"`
	AvgResponseTime      types.Float64 `tfsdk:"avg_response_time"`
	LastExecutionTime    types.String  `tfsdk:"last_execution_time"`
}

func NewPromptResource() resource.Resource {
	return &promptResource{}
}

func (r *promptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (r *promptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge prompt resource",
		Description:         "Manages a ContextForge prompt resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Prompt ID",
				Description:         "Prompt ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Prompt name",
				Description:         "Prompt name",
				Required:            true,
			},
			"original_name": schema.StringAttribute{
				MarkdownDescription: "Original prompt name from the gateway",
				Description:         "Original prompt name from the gateway",
				Computed:            true,
			},
			"custom_name": schema.StringAttribute{
				MarkdownDescription: "Custom prompt name override",
				Description:         "Custom prompt name override",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_name_slug": schema.StringAttribute{
				MarkdownDescription: "Slug generated from custom prompt name",
				Description:         "Slug generated from custom prompt name",
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "User-facing prompt display name",
				Description:         "User-facing prompt display name",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway_slug": schema.StringAttribute{
				MarkdownDescription: "Gateway slug the prompt belongs to",
				Description:         "Gateway slug the prompt belongs to",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Prompt description",
				Description:         "Prompt description",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Prompt template",
				Description:         "Prompt template",
				Required:            true,
			},
			"arguments": schema.ListNestedAttribute{
				MarkdownDescription: "Prompt arguments/parameters",
				Description:         "Prompt arguments/parameters",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Argument name",
							Description:         "Argument name",
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Argument description",
							Description:         "Argument description",
							Optional:            true,
						},
						"required": schema.BoolAttribute{
							MarkdownDescription: "Whether the argument is required",
							Description:         "Whether the argument is required",
							Optional:            true,
						},
					},
				},
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the prompt is active",
				Description:         "Whether the prompt is active",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the prompt is enabled",
				Description:         "Whether the prompt is enabled",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Prompt tags",
				Description:         "Prompt tags",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "Team name",
				Description:         "Team name",
				Computed:            true,
			},
			"owner_email": schema.StringAttribute{
				MarkdownDescription: "Owner email",
				Description:         "Owner email",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility setting",
				Description:         "Visibility setting",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},

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

func (r *promptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *promptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data promptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	promptCreate := &contextforge.PromptCreate{
		Name:     data.Name.ValueString(),
		Template: data.Template.ValueString(),
	}

	if !data.CustomName.IsNull() && !data.CustomName.IsUnknown() {
		customName := data.CustomName.ValueString()
		promptCreate.CustomName = &customName
	}
	if !data.DisplayName.IsNull() && !data.DisplayName.IsUnknown() {
		displayName := data.DisplayName.ValueString()
		promptCreate.DisplayName = &displayName
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		description := data.Description.ValueString()
		promptCreate.Description = &description
	}

	if !data.Arguments.IsNull() && !data.Arguments.IsUnknown() {
		arguments, diags := promptArgumentsFromList(ctx, data.Arguments)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		promptCreate.Arguments = arguments
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		promptCreate.Tags = tags
	}

	opts := &contextforge.PromptCreateOptions{}
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		opts.TeamID = &teamID
	}
	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		opts.Visibility = &visibility
	}

	prompt, _, err := r.client.Prompts.Create(ctx, promptCreate, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Prompt",
			fmt.Sprintf("Unable to create prompt; %v", err),
		)
		return
	}

	r.mapPromptToState(ctx, prompt, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *promptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data promptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prompt, err := r.getPromptByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Prompt",
			fmt.Sprintf("Unable to read prompt with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}
	if prompt == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapPromptToState(ctx, prompt, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *promptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data promptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	template := data.Template.ValueString()
	promptUpdate := &contextforge.PromptUpdate{
		Name:     &name,
		Template: &template,
	}

	if !data.CustomName.IsUnknown() {
		if data.CustomName.IsNull() {
			empty := ""
			promptUpdate.CustomName = &empty
		} else {
			customName := data.CustomName.ValueString()
			promptUpdate.CustomName = &customName
		}
	}

	if !data.DisplayName.IsUnknown() {
		if data.DisplayName.IsNull() {
			empty := ""
			promptUpdate.DisplayName = &empty
		} else {
			displayName := data.DisplayName.ValueString()
			promptUpdate.DisplayName = &displayName
		}
	}

	if !data.Description.IsUnknown() {
		if data.Description.IsNull() {
			empty := ""
			promptUpdate.Description = &empty
		} else {
			description := data.Description.ValueString()
			promptUpdate.Description = &description
		}
	}

	if !data.Arguments.IsUnknown() {
		if data.Arguments.IsNull() {
			promptUpdate.Arguments = []contextforge.PromptArgument{}
		} else {
			arguments, diags := promptArgumentsFromList(ctx, data.Arguments)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			promptUpdate.Arguments = arguments
		}
	}

	if !data.Tags.IsUnknown() {
		if data.Tags.IsNull() {
			promptUpdate.Tags = []string{}
		} else {
			var tags []string
			resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			promptUpdate.Tags = tags
		}
	}

	if !data.TeamID.IsUnknown() {
		if data.TeamID.IsNull() {
			empty := ""
			promptUpdate.TeamID = &empty
		} else {
			teamID := data.TeamID.ValueString()
			promptUpdate.TeamID = &teamID
		}
	}

	if !data.Visibility.IsUnknown() {
		if data.Visibility.IsNull() {
			empty := ""
			promptUpdate.Visibility = &empty
		} else {
			visibility := data.Visibility.ValueString()
			promptUpdate.Visibility = &visibility
		}
	}

	_, httpResp, err := r.client.Prompts.Update(ctx, data.ID.ValueString(), promptUpdate)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Update Prompt",
			fmt.Sprintf("Unable to update prompt with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	updatedPrompt, err := r.getPromptByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Prompt After Update",
			fmt.Sprintf("Unable to read prompt after update; %v", err),
		)
		return
	}
	if updatedPrompt == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapPromptToState(ctx, updatedPrompt, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *promptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data promptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Prompts.Delete(ctx, data.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Prompt",
			fmt.Sprintf("Unable to delete prompt with ID %s; %v", data.ID.ValueString(), err),
		)
	}
}

func (r *promptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *promptResource) getPromptByID(ctx context.Context, promptID string) (*contextforge.Prompt, error) {
	opts := &contextforge.PromptListOptions{IncludeInactive: true}

	for {
		prompts, listResp, err := r.client.Prompts.List(ctx, opts)
		if err != nil {
			return nil, err
		}

		for _, prompt := range prompts {
			if prompt.ID == promptID {
				return prompt, nil
			}
		}

		if listResp == nil || listResp.NextCursor == "" {
			break
		}

		opts.Cursor = listResp.NextCursor
	}

	return nil, nil
}

func (r *promptResource) mapPromptToState(ctx context.Context, prompt *contextforge.Prompt, data *promptResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(prompt.ID)
	if prompt.OriginalName != nil && *prompt.OriginalName != "" {
		data.Name = types.StringValue(*prompt.OriginalName)
	} else {
		data.Name = types.StringValue(prompt.Name)
	}
	data.OriginalName = types.StringPointerValue(prompt.OriginalName)
	data.CustomName = types.StringPointerValue(prompt.CustomName)
	data.CustomNameSlug = types.StringPointerValue(prompt.CustomNameSlug)
	data.DisplayName = types.StringPointerValue(prompt.DisplayName)
	data.GatewaySlug = types.StringPointerValue(prompt.GatewaySlug)
	data.Description = types.StringPointerValue(prompt.Description)
	data.Template = types.StringValue(prompt.Template)
	data.IsActive = types.BoolValue(prompt.IsActive)
	data.Enabled = types.BoolValue(prompt.Enabled || prompt.IsActive)

	if len(prompt.Arguments) > 0 {
		argModels := make([]promptResourceArgumentModel, len(prompt.Arguments))
		for i, argument := range prompt.Arguments {
			argModels[i] = promptResourceArgumentModel{
				Name:        types.StringValue(argument.Name),
				Description: types.StringPointerValue(argument.Description),
				Required:    types.BoolValue(argument.Required),
			}
		}
		argumentsList, argsDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: promptResourceArgumentModel{}.attrTypes()}, argModels)
		diags.Append(argsDiags...)
		data.Arguments = argumentsList
	} else {
		data.Arguments = types.ListNull(types.ObjectType{AttrTypes: promptResourceArgumentModel{}.attrTypes()})
	}

	if prompt.Metrics != nil {
		metricsModel := promptResourceMetricsModel{
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

		metricsObject, metricsDiags := types.ObjectValueFrom(ctx, metricsModel.attrTypes(), metricsModel)
		diags.Append(metricsDiags...)
		if !diags.HasError() {
			data.Metrics = metricsObject
		}
	} else {
		data.Metrics = types.ObjectNull(promptResourceMetricsModel{}.attrTypes())
	}

	if prompt.Tags != nil {
		tagsList, tagsDiags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(prompt.Tags))
		diags.Append(tagsDiags...)
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
}

func promptArgumentsFromList(ctx context.Context, list types.List) ([]contextforge.PromptArgument, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var argModels []promptResourceArgumentModel
	diags.Append(list.ElementsAs(ctx, &argModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	arguments := make([]contextforge.PromptArgument, len(argModels))
	for i, arg := range argModels {
		arguments[i] = contextforge.PromptArgument{
			Name:     arg.Name.ValueString(),
			Required: !arg.Required.IsNull() && arg.Required.ValueBool(),
		}
		if !arg.Description.IsNull() && !arg.Description.IsUnknown() {
			description := arg.Description.ValueString()
			arguments[i].Description = &description
		}
	}

	return arguments, diags
}

func (m promptResourceArgumentModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"required":    types.BoolType,
	}
}

func (m promptResourceMetricsModel) attrTypes() map[string]attr.Type {
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
