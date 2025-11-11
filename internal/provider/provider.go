package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/leefowlercu/go-contextforge/contextforge"
)

// ContextForgeProvider is the provider implementation.
type ContextForgeProvider struct {
	Version string
}

// Force compile-time validation that ContextForgeProvider satisfies the provider.Provider interface.
var _ provider.Provider = &ContextForgeProvider{}

// ContextForgeProviderModel defines the provider-level configuration data model.
type ContextForgeProviderModel struct {
	Address types.String `tfsdk:"address"`
	Token   types.String `tfsdk:"token"`
}

// New is a helper function to that returns a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ContextForgeProvider{
			Version: version,
		}
	}
}

// Metadata returns the provider type name and version.
func (p *ContextForgeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "contextforge"
	resp.Version = p.Version
}

// Schema defines the provider-level schema for configuration data.
func (p *ContextForgeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for IBM ContextForge MCP Gateway management. " +
			"Manages virtual servers, gateways, tools, resources, and prompts for the ContextForge MCP Gateway service.",
		MarkdownDescription: "Terraform provider for **IBM ContextForge MCP Gateway** gateway management. " +
			"Manages virtual servers, gateways, tools, resources, and prompts for the ContextForge MCP Gateway service.\n\n" +
			"See the [ContextForge MCP Gateway documentation](https://github.com/IBM/mcp-context-forge) for more information.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "ContextForge MCP Gateway address URL (e.g., https://contextforge.example.com). " +
					"This is a URL with a scheme, a hostname and a port but with no path. Can also be set via CONTEXTFORGE_ADDR environment variable.",
				MarkdownDescription: "Context Forge gateway address URL (e.g., `https://contextforge.example.com`). " +
					"This is a URL with a scheme, a hostname and a port but with no path. Can also be set via `CONTEXTFORGE_ADDR` environment variable.",
				Optional: true,
			},
			"token": schema.StringAttribute{
				Description: "JWT token used to authenticate with the ContextForge MCP Gateway. " +
					"Can also be set via CONTEXTFORGE_TOKEN environment variable.",
				MarkdownDescription: "JWT token used to authenticate with the ContextForge MCP Gateway. " +
					"Can also be set via `CONTEXTFORGE_TOKEN` environment variable.",
				Sensitive: true,
				Optional:  true,
			},
		},
	}
}

// Configure prepares the ContextForge MCP Gateway client for data sources and resources.
func (p *ContextForgeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring ContextForge client")

	// Read Terraform configuration into model
	var config ContextForgeProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	// Return any accumulated errors
	if resp.Diagnostics.HasError() {
		return
	}

	// If address configuration value was provided, validate that it is not unknown
	if config.Address.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Unknown ContextForge Address",
			"The provider cannot create the ContextForge client as there is an unknown configuration value for the ContextForge address. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CONTEXTFORGE_ADDR environment variable.",
		)
	}

	// If token configuration value was provided, validate that it is not unknown
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown ContextForge Token",
			"The provider cannot create the ContextForge client as there is an unknown configuration value for the ContextForge token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CONTEXTFORGE_TOKEN environment variable.",
		)
	}

	// Return any accumulated errors
	if resp.Diagnostics.HasError() {
		return
	}

	// Start with environment variables as defaults
	address := os.Getenv("CONTEXTFORGE_ADDR")
	token := os.Getenv("CONTEXTFORGE_TOKEN")

	// Override with explicit config values (config takes precedence)
	if !config.Address.IsNull() {
		address = config.Address.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// Validate address value is present from either source
	if address == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Missing ContextForge API Address",
			"The provider cannot create the ContextForge client as the address configuration value is missing. "+
				"Ensure the address is set in the provider configuration block or via the CONTEXTFORGE_ADDR environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Validate token value is present from either source
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing ContextForge API Token",
			"The provider cannot create the ContextForge client as the token configuration value is missing. "+
				"Ensure the token is set in the provider configuration block or via the CONTEXTFORGE_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Return any accumulated errors
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := contextforge.NewClient(nil, address, token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create ContextForge client",
			"An unexpected error occurred when creating the ContextForge client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *ContextForgeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGatewayDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *ContextForgeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}
