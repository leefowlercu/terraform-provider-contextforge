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
	"github.com/leefowlercu/terraform-provider-contextforge/internal/tfconv"
)

type gatewayResource struct {
	client *contextforge.Client
}

// Force compile-time validation
var _ resource.Resource = &gatewayResource{}
var _ resource.ResourceWithConfigure = &gatewayResource{}
var _ resource.ResourceWithImportState = &gatewayResource{}

// gatewayResourceModel defines the resource model (same as data source minus lookup-only semantics)
type gatewayResourceModel struct {
	// Computed field
	ID types.String `tfsdk:"id"`

	// Core fields
	Name         types.String  `tfsdk:"name"`
	URL          types.String  `tfsdk:"url"`
	Description  types.String  `tfsdk:"description"`
	Transport    types.String  `tfsdk:"transport"`
	Enabled      types.Bool    `tfsdk:"enabled"`
	Reachable    types.Bool    `tfsdk:"reachable"`
	Capabilities types.Dynamic `tfsdk:"capabilities"`

	// Authentication fields
	PassthroughHeaders types.List    `tfsdk:"passthrough_headers"`
	AuthType           types.String  `tfsdk:"auth_type"`
	AuthUsername       types.String  `tfsdk:"auth_username"`
	AuthPassword       types.String  `tfsdk:"auth_password"`
	AuthToken          types.String  `tfsdk:"auth_token"`
	AuthHeaderKey      types.String  `tfsdk:"auth_header_key"`
	AuthHeaderValue    types.String  `tfsdk:"auth_header_value"`
	AuthHeaders        types.List    `tfsdk:"auth_headers"`
	AuthValue          types.String  `tfsdk:"auth_value"`
	OAuthConfig        types.Dynamic `tfsdk:"oauth_config"`

	// Organizational fields
	Tags       types.List   `tfsdk:"tags"`
	TeamID     types.String `tfsdk:"team_id"`
	Team       types.String `tfsdk:"team"`
	OwnerEmail types.String `tfsdk:"owner_email"`
	Visibility types.String `tfsdk:"visibility"`

	// Timestamps
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	LastSeen  types.String `tfsdk:"last_seen"`

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
	Slug              types.String `tfsdk:"slug"`
}

// NewGatewayResource is a helper function to instantiate the gateway resource.
func NewGatewayResource() resource.Resource {
	return &gatewayResource{}
}

// Metadata returns the resource type name.
func (r *gatewayResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

// Schema defines the schema for the resource.
func (r *gatewayResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ContextForge gateway resource",
		Description:         "Manages a ContextForge gateway resource",

		Attributes: map[string]schema.Attribute{
			// Computed field
			"id": schema.StringAttribute{
				MarkdownDescription: "Gateway ID",
				Description:         "Gateway ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Core required fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Gateway name",
				Description:         "Gateway name",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Gateway URL endpoint",
				Description:         "Gateway URL endpoint",
				Required:            true,
			},
			"transport": schema.StringAttribute{
				MarkdownDescription: "Transport protocol (SSE, HTTP, etc.)",
				Description:         "Transport protocol (SSE, HTTP, etc.)",
				Required:            true,
			},

			// Core optional fields
			"description": schema.StringAttribute{
				MarkdownDescription: "Gateway description",
				Description:         "Gateway description",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the gateway is enabled (default: true)",
				Description:         "Whether the gateway is enabled (default: true)",
				Optional:            true,
				Computed:            true,
			},

			// Computed runtime status
			"reachable": schema.BoolAttribute{
				MarkdownDescription: "Whether the gateway is currently reachable",
				Description:         "Whether the gateway is currently reachable",
				Computed:            true,
			},
			"capabilities": schema.DynamicAttribute{
				MarkdownDescription: "Gateway capabilities",
				Description:         "Gateway capabilities",
				Computed:            true,
			},

			// Authentication fields
			"passthrough_headers": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Headers to pass through to the gateway",
				Description:         "Headers to pass through to the gateway",
				Optional:            true,
				Computed:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type (basic, bearer, authheaders, oauth)",
				Description:         "Authentication type (basic, bearer, authheaders, oauth)",
				Optional:            true,
			},
			"auth_username": schema.StringAttribute{
				MarkdownDescription: "Username for basic authentication",
				Description:         "Username for basic authentication",
				Optional:            true,
			},
			"auth_password": schema.StringAttribute{
				MarkdownDescription: "Password for basic authentication",
				Description:         "Password for basic authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Token for bearer authentication",
				Description:         "Token for bearer authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_header_key": schema.StringAttribute{
				MarkdownDescription: "Custom auth header key",
				Description:         "Custom auth header key",
				Optional:            true,
			},
			"auth_header_value": schema.StringAttribute{
				MarkdownDescription: "Custom auth header value",
				Description:         "Custom auth header value",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_headers": schema.ListAttribute{
				ElementType:         types.MapType{ElemType: types.StringType},
				MarkdownDescription: "List of authentication headers",
				Description:         "List of authentication headers",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_value": schema.StringAttribute{
				MarkdownDescription: "Raw authentication value",
				Description:         "Raw authentication value",
				Optional:            true,
				Sensitive:           true,
			},
			"oauth_config": schema.DynamicAttribute{
				MarkdownDescription: "OAuth configuration",
				Description:         "OAuth configuration",
				Optional:            true,
				Sensitive:           true,
			},

			// Organizational fields
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Gateway tags",
				Description:         "Gateway tags",
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
				MarkdownDescription: "Owner email",
				Description:         "Owner email",
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
			"last_seen": schema.StringAttribute{
				MarkdownDescription: "Last seen timestamp (RFC3339 format)",
				Description:         "Last seen timestamp (RFC3339 format)",
				Computed:            true,
			},

			// Metadata (all computed/read-only)
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User who created the gateway",
				Description:         "User who created the gateway",
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
				MarkdownDescription: "User who last modified the gateway",
				Description:         "User who last modified the gateway",
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
				MarkdownDescription: "Import batch ID",
				Description:         "Import batch ID",
				Computed:            true,
			},
			"federation_source": schema.StringAttribute{
				MarkdownDescription: "Federation source",
				Description:         "Federation source",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Gateway version",
				Description:         "Gateway version",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Gateway slug",
				Description:         "Gateway slug",
				Computed:            true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data gatewayResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build gateway object for API
	gateway := &contextforge.Gateway{
		Name:      data.Name.ValueString(),
		URL:       data.URL.ValueString(),
		Transport: data.Transport.ValueString(),
	}

	// Map optional description
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		gateway.Description = &desc
	}

	// Map optional enabled (default true handled by API)
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		gateway.Enabled = data.Enabled.ValueBool()
	}

	// Map optional passthrough_headers
	if !data.PassthroughHeaders.IsNull() && !data.PassthroughHeaders.IsUnknown() {
		var headers []string
		resp.Diagnostics.Append(data.PassthroughHeaders.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		gateway.PassthroughHeaders = headers
	}

	// Map authentication fields (these are mutually exclusive based on auth_type)
	if !data.AuthType.IsNull() && !data.AuthType.IsUnknown() {
		authType := data.AuthType.ValueString()
		gateway.AuthType = &authType
	}
	if !data.AuthUsername.IsNull() && !data.AuthUsername.IsUnknown() {
		username := data.AuthUsername.ValueString()
		gateway.AuthUsername = &username
	}
	if !data.AuthPassword.IsNull() && !data.AuthPassword.IsUnknown() {
		password := data.AuthPassword.ValueString()
		gateway.AuthPassword = &password
	}
	if !data.AuthToken.IsNull() && !data.AuthToken.IsUnknown() {
		token := data.AuthToken.ValueString()
		gateway.AuthToken = &token
	}
	if !data.AuthHeaderKey.IsNull() && !data.AuthHeaderKey.IsUnknown() {
		key := data.AuthHeaderKey.ValueString()
		gateway.AuthHeaderKey = &key
	}
	if !data.AuthHeaderValue.IsNull() && !data.AuthHeaderValue.IsUnknown() {
		value := data.AuthHeaderValue.ValueString()
		gateway.AuthHeaderValue = &value
	}

	// Map auth_headers (list of maps)
	if !data.AuthHeaders.IsNull() && !data.AuthHeaders.IsUnknown() {
		var authHeadersList []types.Map
		resp.Diagnostics.Append(data.AuthHeaders.ElementsAs(ctx, &authHeadersList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var authHeaders []map[string]string
		for _, headerMap := range authHeadersList {
			var headerMapData map[string]string
			resp.Diagnostics.Append(headerMap.ElementsAs(ctx, &headerMapData, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			authHeaders = append(authHeaders, headerMapData)
		}
		gateway.AuthHeaders = authHeaders
	}

	// Map auth_value
	if !data.AuthValue.IsNull() && !data.AuthValue.IsUnknown() {
		authValue := data.AuthValue.ValueString()
		gateway.AuthValue = &authValue
	}

	// Map oauth_config (Dynamic type)
	if !data.OAuthConfig.IsNull() && !data.OAuthConfig.IsUnknown() {
		oauthMap, err := tfconv.ConvertObjectValueToMap(ctx, data.OAuthConfig.UnderlyingValue())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert OAuth Config",
				fmt.Sprintf("Unable to convert oauth_config from object value; %v", err),
			)
			return
		}
		gateway.OAuthConfig = oauthMap
	}

	// Map optional tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		gateway.Tags = contextforge.NewTags(tags)
	}

	// Prepare create options for team_id and visibility (like Tool resource)
	opts := &contextforge.GatewayCreateOptions{}

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

	// Create gateway via API
	createdGateway, _, err := r.client.Gateways.Create(ctx, gateway, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Gateway",
			fmt.Sprintf("Unable to create gateway; %v", err),
		)
		return
	}

	// Map response to state using helper
	r.mapGatewayToState(ctx, createdGateway, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data gatewayResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get gateway from API
	gateway, httpResp, err := r.client.Gateways.Get(ctx, data.ID.ValueString())
	if err != nil {
		// Handle 404 - resource no longer exists
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read Gateway",
			fmt.Sprintf("Unable to read gateway with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map response to state using helper
	r.mapGatewayToState(ctx, gateway, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data gatewayResourceModel

	// Read plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build gateway object for API (similar to Create but for Update endpoint)
	gateway := &contextforge.Gateway{
		Name:      data.Name.ValueString(),
		URL:       data.URL.ValueString(),
		Transport: data.Transport.ValueString(),
	}

	// Map optional fields (same pattern as Create)
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		gateway.Description = &desc
	}

	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		gateway.Enabled = data.Enabled.ValueBool()
	}

	if !data.PassthroughHeaders.IsNull() && !data.PassthroughHeaders.IsUnknown() {
		var headers []string
		resp.Diagnostics.Append(data.PassthroughHeaders.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		gateway.PassthroughHeaders = headers
	}

	// Auth fields
	if !data.AuthType.IsNull() && !data.AuthType.IsUnknown() {
		authType := data.AuthType.ValueString()
		gateway.AuthType = &authType
	}
	if !data.AuthUsername.IsNull() && !data.AuthUsername.IsUnknown() {
		username := data.AuthUsername.ValueString()
		gateway.AuthUsername = &username
	}
	if !data.AuthPassword.IsNull() && !data.AuthPassword.IsUnknown() {
		password := data.AuthPassword.ValueString()
		gateway.AuthPassword = &password
	}
	if !data.AuthToken.IsNull() && !data.AuthToken.IsUnknown() {
		token := data.AuthToken.ValueString()
		gateway.AuthToken = &token
	}
	if !data.AuthHeaderKey.IsNull() && !data.AuthHeaderKey.IsUnknown() {
		key := data.AuthHeaderKey.ValueString()
		gateway.AuthHeaderKey = &key
	}
	if !data.AuthHeaderValue.IsNull() && !data.AuthHeaderValue.IsUnknown() {
		value := data.AuthHeaderValue.ValueString()
		gateway.AuthHeaderValue = &value
	}

	// Auth headers
	if !data.AuthHeaders.IsNull() && !data.AuthHeaders.IsUnknown() {
		var authHeadersList []types.Map
		resp.Diagnostics.Append(data.AuthHeaders.ElementsAs(ctx, &authHeadersList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var authHeaders []map[string]string
		for _, headerMap := range authHeadersList {
			var headerMapData map[string]string
			resp.Diagnostics.Append(headerMap.ElementsAs(ctx, &headerMapData, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			authHeaders = append(authHeaders, headerMapData)
		}
		gateway.AuthHeaders = authHeaders
	}

	if !data.AuthValue.IsNull() && !data.AuthValue.IsUnknown() {
		authValue := data.AuthValue.ValueString()
		gateway.AuthValue = &authValue
	}

	// OAuth config
	if !data.OAuthConfig.IsNull() && !data.OAuthConfig.IsUnknown() {
		oauthMap, err := tfconv.ConvertObjectValueToMap(ctx, data.OAuthConfig.UnderlyingValue())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert OAuth Config",
				fmt.Sprintf("Unable to convert oauth_config from object value; %v", err),
			)
			return
		}
		gateway.OAuthConfig = oauthMap
	}

	// Tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		gateway.Tags = contextforge.NewTags(tags)
	}

	// Team/Visibility (Update uses gateway struct directly, not options)
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueString()
		gateway.TeamID = &teamID
	}

	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		visibility := data.Visibility.ValueString()
		gateway.Visibility = &visibility
	}

	// Update gateway via API
	updatedGateway, httpResp, err := r.client.Gateways.Update(ctx, data.ID.ValueString(), gateway)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Gateway",
			fmt.Sprintf("Unable to update gateway with ID %s (status: %d); %v", data.ID.ValueString(), httpResp.StatusCode, err),
		)
		return
	}

	// Map response to state using helper
	r.mapGatewayToState(ctx, updatedGateway, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data gatewayResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete gateway via API
	httpResp, err := r.client.Gateways.Delete(ctx, data.ID.ValueString())
	if err != nil {
		// Ignore 404 errors (resource already deleted)
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Delete Gateway",
			fmt.Sprintf("Unable to delete gateway with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// State is automatically removed by the framework
}

// Configure adds the provider configured client to the resource.
func (r *gatewayResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID from the import request as the gateway ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapGatewayToState is a helper to map Gateway API response to Terraform state
func (r *gatewayResource) mapGatewayToState(ctx context.Context, gateway *contextforge.Gateway, data *gatewayResourceModel, diags *diag.Diagnostics) {
	// Core fields
	data.ID = types.StringPointerValue(gateway.ID)
	data.Name = types.StringValue(gateway.Name)
	data.URL = types.StringValue(gateway.URL)
	data.Description = types.StringPointerValue(gateway.Description)
	data.Transport = types.StringValue(gateway.Transport)
	data.Enabled = types.BoolValue(gateway.Enabled)
	data.Reachable = types.BoolValue(gateway.Reachable)

	// Capabilities (Dynamic type)
	if gateway.Capabilities != nil {
		capValue, err := tfconv.ConvertMapToObjectValue(ctx, gateway.Capabilities)
		if err != nil {
			diags.AddError(
				"Failed to Convert Capabilities",
				fmt.Sprintf("Unable to convert capabilities to object value; %v", err),
			)
			return
		}
		data.Capabilities = types.DynamicValue(capValue)
	} else {
		data.Capabilities = types.DynamicNull()
	}

	// Passthrough headers
	if gateway.PassthroughHeaders != nil {
		headersList, diagsList := types.ListValueFrom(ctx, types.StringType, gateway.PassthroughHeaders)
		diags.Append(diagsList...)
		data.PassthroughHeaders = headersList
	} else {
		data.PassthroughHeaders = types.ListNull(types.StringType)
	}

	// Auth fields
	data.AuthType = types.StringPointerValue(gateway.AuthType)
	data.AuthUsername = types.StringPointerValue(gateway.AuthUsername)
	data.AuthPassword = types.StringPointerValue(gateway.AuthPassword)
	data.AuthToken = types.StringPointerValue(gateway.AuthToken)
	data.AuthHeaderKey = types.StringPointerValue(gateway.AuthHeaderKey)
	data.AuthHeaderValue = types.StringPointerValue(gateway.AuthHeaderValue)

	// Auth headers (list of maps)
	if gateway.AuthHeaders != nil {
		var authHeadersList []attr.Value
		for _, header := range gateway.AuthHeaders {
			headerMap, diagsList := types.MapValueFrom(ctx, types.StringType, header)
			diags.Append(diagsList...)
			authHeadersList = append(authHeadersList, headerMap)
		}
		authHeadersListValue, diagsList := types.ListValue(types.MapType{ElemType: types.StringType}, authHeadersList)
		diags.Append(diagsList...)
		data.AuthHeaders = authHeadersListValue
	} else {
		data.AuthHeaders = types.ListNull(types.MapType{ElemType: types.StringType})
	}

	data.AuthValue = types.StringPointerValue(gateway.AuthValue)

	// OAuth config (Dynamic type)
	if gateway.OAuthConfig != nil {
		oauthValue, err := tfconv.ConvertMapToObjectValue(ctx, gateway.OAuthConfig)
		if err != nil {
			diags.AddError(
				"Failed to Convert OAuth Config",
				fmt.Sprintf("Unable to convert oauth_config to object value; %v", err),
			)
			return
		}
		data.OAuthConfig = types.DynamicValue(oauthValue)
	} else {
		data.OAuthConfig = types.DynamicNull()
	}

	// Tags
	if gateway.Tags != nil {
		tagsList, diagsList := types.ListValueFrom(ctx, types.StringType, contextforge.TagNames(gateway.Tags))
		diags.Append(diagsList...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	// Organizational fields
	data.TeamID = types.StringPointerValue(gateway.TeamID)
	data.Team = types.StringPointerValue(gateway.Team)
	data.OwnerEmail = types.StringPointerValue(gateway.OwnerEmail)
	data.Visibility = types.StringPointerValue(gateway.Visibility)

	// Timestamps
	if gateway.CreatedAt != nil && !gateway.CreatedAt.Time.IsZero() {
		data.CreatedAt = types.StringValue(gateway.CreatedAt.Time.Format(time.RFC3339))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if gateway.UpdatedAt != nil && !gateway.UpdatedAt.Time.IsZero() {
		data.UpdatedAt = types.StringValue(gateway.UpdatedAt.Time.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	if gateway.LastSeen != nil && !gateway.LastSeen.Time.IsZero() {
		data.LastSeen = types.StringValue(gateway.LastSeen.Time.Format(time.RFC3339))
	} else {
		data.LastSeen = types.StringNull()
	}

	// Metadata fields
	data.CreatedBy = types.StringPointerValue(gateway.CreatedBy)
	data.CreatedFromIP = types.StringPointerValue(gateway.CreatedFromIP)
	data.CreatedVia = types.StringPointerValue(gateway.CreatedVia)
	data.CreatedUserAgent = types.StringPointerValue(gateway.CreatedUserAgent)
	data.ModifiedBy = types.StringPointerValue(gateway.ModifiedBy)
	data.ModifiedFromIP = types.StringPointerValue(gateway.ModifiedFromIP)
	data.ModifiedVia = types.StringPointerValue(gateway.ModifiedVia)
	data.ModifiedUserAgent = types.StringPointerValue(gateway.ModifiedUserAgent)
	data.ImportBatchID = types.StringPointerValue(gateway.ImportBatchID)
	data.FederationSource = types.StringPointerValue(gateway.FederationSource)

	if gateway.Version != nil {
		data.Version = types.Int64Value(int64(*gateway.Version))
	} else {
		data.Version = types.Int64Null()
	}

	data.Slug = types.StringPointerValue(gateway.Slug)
}
