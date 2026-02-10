package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leefowlercu/go-contextforge/contextforge"
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type teamResource struct {
	client *contextforge.Client
}

var _ resource.Resource = &teamResource{}
var _ resource.ResourceWithConfigure = &teamResource{}
var _ resource.ResourceWithImportState = &teamResource{}

type teamResourceModel struct {
	ID types.String `tfsdk:"id"`

	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	Visibility  types.String `tfsdk:"visibility"`
	MaxMembers  types.Int64  `tfsdk:"max_members"`

	IsPersonal  types.Bool   `tfsdk:"is_personal"`
	MemberCount types.Int64  `tfsdk:"member_count"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	CreatedBy   types.String `tfsdk:"created_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge team resource",
		Description:         "Manages a ContextForge team resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Description:         "Team ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Team name",
				Description:         "Team name",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Team slug",
				Description:         "Team slug",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Team description",
				Description:         "Team description",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Team visibility",
				Description:         "Team visibility",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_members": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of members",
				Description:         "Maximum number of members",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_personal": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a personal team",
				Description:         "Whether this is a personal team",
				Computed:            true,
			},
			"member_count": schema.Int64Attribute{
				MarkdownDescription: "Current number of members",
				Description:         "Current number of members",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the team is active",
				Description:         "Whether the team is active",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Creator email",
				Description:         "Creator email",
				Computed:            true,
			},
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
		},
	}
}

func (r *teamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamCreate := &contextforge.TeamCreate{
		Name: data.Name.ValueString(),
	}

	if !data.Slug.IsNull() && !data.Slug.IsUnknown() {
		slug := data.Slug.ValueString()
		teamCreate.Slug = &slug
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		description := data.Description.ValueString()
		teamCreate.Description = &description
	}
	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		teamCreate.Visibility = &visibility
	}
	if !data.MaxMembers.IsNull() && !data.MaxMembers.IsUnknown() {
		maxMembers := int(data.MaxMembers.ValueInt64())
		teamCreate.MaxMembers = &maxMembers
	}

	team, _, err := r.client.Teams.Create(ctx, teamCreate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Team",
			fmt.Sprintf("Unable to create team; %v", err),
		)
		return
	}

	r.mapTeamToState(team, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, httpResp, err := r.client.Teams.Get(ctx, data.ID.ValueString())
	if err != nil && httpResp != nil && (httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden) {
		team, err = findTeamByIDViaList(ctx, r.client, data.ID.ValueString())
	}
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read Team",
			fmt.Sprintf("Unable to read team with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}
	if team == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapTeamToState(team, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	teamUpdate := &contextforge.TeamUpdate{
		Name: &name,
	}

	if !data.Description.IsUnknown() {
		if data.Description.IsNull() {
			empty := ""
			teamUpdate.Description = &empty
		} else {
			description := data.Description.ValueString()
			teamUpdate.Description = &description
		}
	}

	if !data.Visibility.IsUnknown() {
		if data.Visibility.IsNull() {
			empty := ""
			teamUpdate.Visibility = &empty
		} else {
			visibility := data.Visibility.ValueString()
			teamUpdate.Visibility = &visibility
		}
	}

	if !data.MaxMembers.IsNull() && !data.MaxMembers.IsUnknown() {
		maxMembers := int(data.MaxMembers.ValueInt64())
		teamUpdate.MaxMembers = &maxMembers
	}

	team, httpResp, err := r.client.Teams.Update(ctx, data.ID.ValueString(), teamUpdate)
	if err != nil && httpResp != nil && (httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden) {
		team, httpResp, err = updateTeamNoSlash(ctx, r.client, data.ID.ValueString(), teamUpdate)
	}
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Update Team",
			fmt.Sprintf("Unable to update team with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	r.mapTeamToState(team, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Teams.Delete(ctx, data.ID.ValueString())
	if err != nil && httpResp != nil && (httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden) {
		httpResp, err = deleteTeamNoSlash(ctx, r.client, data.ID.ValueString())
	}
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Team",
			fmt.Sprintf("Unable to delete team with ID %s; %v", data.ID.ValueString(), err),
		)
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamResource) mapTeamToState(team *contextforge.Team, data *teamResourceModel) {
	data.ID = types.StringValue(team.ID)
	data.Name = types.StringValue(team.Name)
	data.Slug = types.StringValue(team.Slug)
	data.Description = types.StringPointerValue(team.Description)
	data.Visibility = types.StringPointerValue(team.Visibility)
	data.IsPersonal = types.BoolValue(team.IsPersonal)
	data.MemberCount = types.Int64Value(int64(team.MemberCount))
	data.IsActive = types.BoolValue(team.IsActive)
	data.CreatedBy = types.StringValue(team.CreatedBy)

	if team.MaxMembers != nil {
		data.MaxMembers = types.Int64PointerValue(tfconv.Int64Ptr(*team.MaxMembers))
	} else {
		data.MaxMembers = types.Int64Null()
	}

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
}
