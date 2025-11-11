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

type gatewayDataSource struct {
	client *contextforge.Client
}

// Force compile-time validation that gatewayDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &gatewayDataSource{}

// Force compile-time validation that gatewayDataSource satisfies the datasource.DataSourceWithConfigure interface.
var _ datasource.DataSourceWithConfigure = &gatewayDataSource{}

// gatewayDataSourceModel defines the data source model.
type gatewayDataSourceModel struct {
	// Lookup field
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

// NewGatewayDataSource is a helper function to instantiate the gateway data source.
func NewGatewayDataSource() datasource.DataSource {
	return &gatewayDataSource{}
}

// Metadata returns the data source type name.
func (d *gatewayDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

// Schema defines the schema for the data source.
func (d *gatewayDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for looking up a ContextForge gateway by ID",
		Description:         "Data source for looking up a ContextForge gateway by ID",

		Attributes: map[string]schema.Attribute{
			// Lookup field
			"id": schema.StringAttribute{
				MarkdownDescription: "Gateway ID for lookup",
				Description:         "Gateway ID for lookup",
				Required:            true,
			},

			// Core fields
			"name": schema.StringAttribute{
				MarkdownDescription: "Gateway name",
				Description:         "Gateway name",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Gateway URL endpoint",
				Description:         "Gateway URL endpoint",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Gateway description",
				Description:         "Gateway description",
				Computed:            true,
			},
			"transport": schema.StringAttribute{
				MarkdownDescription: "Transport protocol (e.g., SSE, HTTP)",
				Description:         "Transport protocol (e.g., SSE, HTTP)",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the gateway is enabled",
				Description:         "Whether the gateway is enabled",
				Computed:            true,
			},
			"reachable": schema.BoolAttribute{
				MarkdownDescription: "Whether the gateway is reachable",
				Description:         "Whether the gateway is reachable",
				Computed:            true,
			},
			"capabilities": schema.DynamicAttribute{
				MarkdownDescription: "Gateway capabilities as a dynamic object",
				Description:         "Gateway capabilities as a dynamic object",
				Computed:            true,
			},

			// Authentication fields
			"passthrough_headers": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Headers to pass through to the gateway",
				Description:         "Headers to pass through to the gateway",
				Computed:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type",
				Description:         "Authentication type",
				Computed:            true,
			},
			"auth_username": schema.StringAttribute{
				MarkdownDescription: "Authentication username",
				Description:         "Authentication username",
				Computed:            true,
			},
			"auth_password": schema.StringAttribute{
				MarkdownDescription: "Authentication password",
				Description:         "Authentication password",
				Computed:            true,
				Sensitive:           true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Authentication token",
				Description:         "Authentication token",
				Computed:            true,
				Sensitive:           true,
			},
			"auth_header_key": schema.StringAttribute{
				MarkdownDescription: "Custom authentication header key",
				Description:         "Custom authentication header key",
				Computed:            true,
			},
			"auth_header_value": schema.StringAttribute{
				MarkdownDescription: "Custom authentication header value",
				Description:         "Custom authentication header value",
				Computed:            true,
				Sensitive:           true,
			},
			"auth_headers": schema.ListAttribute{
				ElementType:         types.MapType{ElemType: types.StringType},
				MarkdownDescription: "List of authentication header maps",
				Description:         "List of authentication header maps",
				Computed:            true,
				Sensitive:           true,
			},
			"auth_value": schema.StringAttribute{
				MarkdownDescription: "Generic authentication value",
				Description:         "Generic authentication value",
				Computed:            true,
				Sensitive:           true,
			},
			"oauth_config": schema.DynamicAttribute{
				MarkdownDescription: "OAuth configuration as a dynamic object",
				Description:         "OAuth configuration as a dynamic object",
				Computed:            true,
				Sensitive:           true,
			},

			// Organizational fields
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Gateway tags",
				Description:         "Gateway tags",
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
			"last_seen": schema.StringAttribute{
				MarkdownDescription: "Last seen timestamp (RFC3339 format)",
				Description:         "Last seen timestamp (RFC3339 format)",
				Computed:            true,
			},

			// Metadata (read-only)
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User who created the gateway (read-only metadata)",
				Description:         "User who created the gateway (read-only metadata)",
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
				MarkdownDescription: "User who last modified the gateway (read-only metadata)",
				Description:         "User who last modified the gateway (read-only metadata)",
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
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly slug (read-only metadata)",
				Description:         "URL-friendly slug (read-only metadata)",
				Computed:            true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data gatewayDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ID is provided
	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up a gateway",
		)
		return
	}

	// Get gateway from API
	gateway, _, err := d.client.Gateways.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Gateway",
			fmt.Sprintf("Unable to read gateway with ID %s; %v", data.ID.ValueString(), err),
		)
		return
	}

	// Map core fields
	data.ID = types.StringPointerValue(gateway.ID)
	data.Name = types.StringValue(gateway.Name)
	data.URL = types.StringValue(gateway.URL)
	data.Description = types.StringPointerValue(gateway.Description)
	data.Transport = types.StringValue(gateway.Transport)
	data.Enabled = types.BoolValue(gateway.Enabled)
	data.Reachable = types.BoolValue(gateway.Reachable)

	// Map capabilities (map[string]any -> types.Dynamic)
	if gateway.Capabilities != nil {
		capValue, err := tfconv.ConvertMapToObjectValue(ctx, gateway.Capabilities)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert Capabilities",
				fmt.Sprintf("Unable to convert capabilities to object value; %v", err),
			)
			return
		}
		data.Capabilities = types.DynamicValue(capValue)
	} else {
		data.Capabilities = types.DynamicNull()
	}

	// Map authentication fields
	if gateway.PassthroughHeaders != nil {
		passthroughList, diags := types.ListValueFrom(ctx, types.StringType, gateway.PassthroughHeaders)
		resp.Diagnostics.Append(diags...)
		data.PassthroughHeaders = passthroughList
	} else {
		data.PassthroughHeaders = types.ListNull(types.StringType)
	}

	data.AuthType = types.StringPointerValue(gateway.AuthType)
	data.AuthUsername = types.StringPointerValue(gateway.AuthUsername)
	data.AuthPassword = types.StringPointerValue(gateway.AuthPassword)
	data.AuthToken = types.StringPointerValue(gateway.AuthToken)
	data.AuthHeaderKey = types.StringPointerValue(gateway.AuthHeaderKey)
	data.AuthHeaderValue = types.StringPointerValue(gateway.AuthHeaderValue)

	// Map auth_headers ([]map[string]string -> types.List of types.Map)
	if gateway.AuthHeaders != nil {
		var authHeadersList []attr.Value
		for _, header := range gateway.AuthHeaders {
			headerMap, diags := types.MapValueFrom(ctx, types.StringType, header)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			authHeadersList = append(authHeadersList, headerMap)
		}
		authHeadersListValue, diags := types.ListValue(types.MapType{ElemType: types.StringType}, authHeadersList)
		resp.Diagnostics.Append(diags...)
		data.AuthHeaders = authHeadersListValue
	} else {
		data.AuthHeaders = types.ListNull(types.MapType{ElemType: types.StringType})
	}

	data.AuthValue = types.StringPointerValue(gateway.AuthValue)

	// Map oauth_config (map[string]any -> types.Dynamic)
	if gateway.OAuthConfig != nil {
		oauthValue, err := tfconv.ConvertMapToObjectValue(ctx, gateway.OAuthConfig)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Convert OAuth Config",
				fmt.Sprintf("Unable to convert oauth_config to object value; %v", err),
			)
			return
		}
		data.OAuthConfig = types.DynamicValue(oauthValue)
	} else {
		data.OAuthConfig = types.DynamicNull()
	}

	// Map organizational fields
	if gateway.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, gateway.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	data.TeamID = types.StringPointerValue(gateway.TeamID)
	data.Team = types.StringPointerValue(gateway.Team)
	data.OwnerEmail = types.StringPointerValue(gateway.OwnerEmail)
	data.Visibility = types.StringPointerValue(gateway.Visibility)

	// Map timestamps (convert to RFC3339 string)
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

	// Map metadata fields
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
	data.Slug = types.StringPointerValue(gateway.Slug)

	if gateway.Version != nil {
		data.Version = types.Int64PointerValue(tfconv.Int64Ptr(*gateway.Version))
	} else {
		data.Version = types.Int64Null()
	}

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *gatewayDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
