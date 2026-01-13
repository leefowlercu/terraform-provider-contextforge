package provider

import (
	"context"
	"fmt"
	"time"

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

type resourceResource struct {
	client *contextforge.Client
}

// Force compile-time validation that resourceResource satisfies the resource.Resource interface.
var _ resource.Resource = &resourceResource{}

// Force compile-time validation that resourceResource satisfies the resource.ResourceWithConfigure interface.
var _ resource.ResourceWithConfigure = &resourceResource{}

// Force compile-time validation that resourceResource satisfies the resource.ResourceWithImportState interface.
var _ resource.ResourceWithImportState = &resourceResource{}

// resourceResourceModel defines the resource model.
type resourceResourceModel struct {
	// Core fields
	ID          types.String `tfsdk:"id"`
	URI         types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	Content     types.String `tfsdk:"content"`
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

// NewResourceResource is a helper function to instantiate the resource resource.
func NewResourceResource() resource.Resource {
	return &resourceResource{}
}

// Metadata returns the resource type name.
func (r *resourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

// Schema defines the schema for the resource.
func (r *resourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge resource",
		Description:         "Manages a ContextForge resource",

		Attributes: map[string]schema.Attribute{
			// Core fields
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource unique identifier",
				Description:         "Resource unique identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Resource URI (required for creation)",
				Description:         "Resource URI (required for creation)",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name (required for creation)",
				Description:         "Resource name (required for creation)",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Resource content (required for creation, not returned by API)",
				Description:         "Resource content (required for creation, not returned by API)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Resource description",
				Description:         "Resource description",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "Resource MIME type",
				Description:         "Resource MIME type",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Resource size in bytes (read-only, computed by backend)",
				Description:         "Resource size in bytes (read-only, computed by backend)",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the resource is active (read-only, computed by backend)",
				Description:         "Whether the resource is active (read-only, computed by backend)",
				Computed:            true,
			},

			// Nested metrics
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "Resource performance metrics (read-only)",
				Description:         "Resource performance metrics (read-only)",
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
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID (can only be set at creation)",
				Description:         "Team ID (can only be set at creation)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "Team name (read-only)",
				Description:         "Team name (read-only)",
				Computed:            true,
			},
			"owner_email": schema.StringAttribute{
				MarkdownDescription: "Owner email address (read-only)",
				Description:         "Owner email address (read-only)",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility setting (public, private, etc.) - can only be set at creation",
				Description:         "Visibility setting (public, private, etc.) - can only be set at creation",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Timestamps (read-only)
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp (RFC3339 format, read-only)",
				Description:         "Creation timestamp (RFC3339 format, read-only)",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp (RFC3339 format, read-only)",
				Description:         "Last update timestamp (RFC3339 format, read-only)",
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

// Configure adds the provider configured client to the resource.
func (r *resourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *resourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the resource create request
	resourceCreate := &contextforge.ResourceCreate{
		URI:     data.URI.ValueString(),
		Name:    data.Name.ValueString(),
		Content: data.Content.ValueString(),
	}

	// Optional fields
	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		resourceCreate.Description = &desc
	}

	if !data.MimeType.IsNull() {
		mimeType := data.MimeType.ValueString()
		resourceCreate.MimeType = &mimeType
	}

	// Note: Size and IsActive are not supported by ResourceCreate API
	// They are computed fields managed by the backend

	// Tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resourceCreate.Tags = tags
	}

	// CreateOptions for team_id and visibility
	opts := &contextforge.ResourceCreateOptions{}
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		opts.TeamID = &teamID
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		opts.Visibility = &visibility
	}

	// Create the resource
	createdResource, _, err := r.client.Resources.Create(ctx, resourceCreate, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Resource",
			fmt.Sprintf("Unable to create resource; %v", err),
		)
		return
	}

	// Map response to state
	r.mapResourceToState(ctx, createdResource, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *resourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get resource from API using List and filter
	// Note: The API doesn't have a dedicated Get endpoint for resources by ID,
	// so we must use List() and filter by ID
	resources, _, err := r.client.Resources.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to List Resources",
			fmt.Sprintf("Unable to list resources; %v", err),
		)
		return
	}

	// Find the resource by ID
	var foundResource *contextforge.Resource
	targetID := data.ID.ValueString()
	for _, res := range resources {
		if res.ID != nil && res.ID.String() == targetID {
			foundResource = res
			break
		}
	}

	if foundResource == nil {
		resp.Diagnostics.AddError(
			"Resource Not Found",
			fmt.Sprintf("Unable to find resource with ID %s", targetID),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to state
	r.mapResourceToState(ctx, foundResource, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *resourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resourceResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the resource update request using three-state semantics
	resourceUpdate := &contextforge.ResourceUpdate{}

	// URI (required field, always set)
	if !data.URI.IsNull() && !data.URI.IsUnknown() {
		uri := data.URI.ValueString()
		resourceUpdate.URI = &uri
	}

	// Name (required field, always set)
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		resourceUpdate.Name = &name
	}

	// Optional fields use three-state semantics
	if !data.Description.IsUnknown() {
		if data.Description.IsNull() {
			emptyStr := ""
			resourceUpdate.Description = &emptyStr // Explicitly clear
		} else {
			desc := data.Description.ValueString()
			resourceUpdate.Description = &desc
		}
	}

	if !data.MimeType.IsUnknown() {
		if data.MimeType.IsNull() {
			emptyStr := ""
			resourceUpdate.MimeType = &emptyStr // Explicitly clear
		} else {
			mimeType := data.MimeType.ValueString()
			resourceUpdate.MimeType = &mimeType
		}
	}

	// Note: Size and IsActive are not supported by ResourceUpdate API
	// They are computed fields managed by the backend

	// Tags
	if !data.Tags.IsUnknown() {
		if data.Tags.IsNull() {
			resourceUpdate.Tags = []string{} // Clear tags
		} else {
			var tags []string
			diags := data.Tags.ElementsAs(ctx, &tags, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resourceUpdate.Tags = tags
		}
	}

	// Get the resource ID
	resourceID := data.ID.ValueString()

	// Update the resource
	_, _, err := r.client.Resources.Update(ctx, resourceID, resourceUpdate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Resource",
			fmt.Sprintf("Unable to update resource; %v", err),
		)
		return
	}

	// The Update API response doesn't include all fields (e.g., team_id is null).
	// Workaround: Do a fresh GET via List and filter to get complete state.
	resources, _, err := r.client.Resources.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Resource After Update",
			fmt.Sprintf("Unable to read resource after update; %v", err),
		)
		return
	}

	// Find the updated resource by ID
	var updatedResource *contextforge.Resource
	for _, res := range resources {
		if res.ID != nil && res.ID.String() == resourceID {
			updatedResource = res
			break
		}
	}

	if updatedResource == nil {
		resp.Diagnostics.AddError(
			"Resource Not Found After Update",
			fmt.Sprintf("Unable to find resource with ID %s after update", resourceID),
		)
		return
	}

	// Map response to state
	r.mapResourceToState(ctx, updatedResource, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *resourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	resourceID := data.ID.ValueString()
	_, err := r.client.Resources.Delete(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Resource",
			fmt.Sprintf("Unable to delete resource; %v", err),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *resourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapResourceToState maps SDK Resource to Terraform state model.
func (r *resourceResource) mapResourceToState(ctx context.Context, resource *contextforge.Resource, data *resourceResourceModel, diags *diag.Diagnostics) {
	// Map core fields
	// Note: resource.ID is *FlexibleID, need to use .String() method
	if resource.ID != nil {
		data.ID = types.StringValue(resource.ID.String())
	}
	data.URI = types.StringValue(resource.URI)
	data.Name = types.StringValue(resource.Name)
	// Note: Content is not returned by API - keep existing state value
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
		metricsObject, metricsDiags := types.ObjectValueFrom(ctx, metricsModel.attrTypes(), metricsModel)
		diags.Append(metricsDiags...)
		if diags.HasError() {
			return
		}
		data.Metrics = metricsObject
	} else {
		data.Metrics = types.ObjectNull(resourceMetricsModel{}.attrTypes())
	}

	// Map organizational fields
	if resource.Tags != nil {
		tagsList, tagsDiags := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(resource.Tags))
		diags.Append(tagsDiags...)
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

	// Convert *int to types.Int64 for version
	if resource.Version != nil {
		data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*resource.Version))
	} else {
		data.Version = types.Int64Null()
	}
}
