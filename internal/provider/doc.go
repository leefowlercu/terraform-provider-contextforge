// Package provider implements the Terraform provider for IBM ContextForge MCP Gateway.
//
// This package contains the core provider implementation, all data sources, all resources,
// and their associated acceptance tests. The provider communicates with the ContextForge
// MCP Gateway API via the go-contextforge client library.
//
// # Provider Configuration
//
// The provider supports dual-source configuration for address and token:
//
//   - Environment variables (CONTEXTFORGE_ADDR, CONTEXTFORGE_TOKEN) provide defaults
//   - HCL configuration attributes override environment variables
//   - Validation occurs in two phases: unknown value detection and empty value validation
//
// Configuration example:
//
//	provider "contextforge" {
//	  address = "https://contextforge.example.com"
//	  token   = var.contextforge_token
//	}
//
// The Configure() method creates a contextforge.Client and stores it in both
// resp.DataSourceData and resp.ResourceData for downstream data sources and resources.
//
// # Data Source Implementation Pattern
//
// Data sources allow Terraform configurations to fetch information about existing
// ContextForge resources without managing them. Each data source follows this pattern:
//
//  1. Define a struct implementing datasource.DataSource and datasource.DataSourceWithConfigure
//  2. Define a model struct with tfsdk tags mapping to the schema
//  3. Implement required methods: Metadata, Schema, Read, Configure
//  4. Register in provider's DataSources() method
//
// Example structure:
//
//	type gatewayDataSource struct {
//	    client *contextforge.Client
//	}
//
//	type gatewayDataSourceModel struct {
//	    ID          types.String `tfsdk:"id"`
//	    Name        types.String `tfsdk:"name"`
//	    Description types.String `tfsdk:"description"`
//	}
//
//	func NewGatewayDataSource() datasource.DataSource {
//	    return &gatewayDataSource{}
//	}
//
// # Schema Conventions
//
// Data source and resource schemas follow these conventions:
//
//   - Lookup attributes: Required or Optional for querying (data sources only)
//   - Computed attributes: Read-only values from API
//   - Sensitive attributes: Mark secrets as Sensitive: true in schema
//   - Documentation: Provide both Description and MarkdownDescription
//   - Naming: Use snake_case for attribute names matching API responses
//
// # Type Conversions
//
// Use the internal/tfconv package for converting API types to Terraform types:
//
//   - map[string]any → tfconv.ConvertMapToObjectValue() then types.DynamicValue()
//   - []string → types.ListValueFrom(ctx, types.StringType, slice)
//   - *string → types.StringPointerValue(ptr)
//   - int → tfconv.Int64Ptr(val) then types.Int64PointerValue()
//   - time.Time → types.StringValue(timestamp.Format(time.RFC3339))
//
// # Read Method Pattern
//
// The Read method for data sources should:
//
//  1. Parse configuration into model struct
//  2. Validate required lookup attributes
//  3. Call appropriate SDK method to fetch data
//  4. Map API response to model with proper type conversions
//  5. Handle nil/null values appropriately
//  6. Set state with mapped data
//
// Example:
//
//	func (d *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
//	    var data gatewayDataSourceModel
//
//	    // Read configuration
//	    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
//	    if resp.Diagnostics.HasError() {
//	        return
//	    }
//
//	    // Fetch from API
//	    gateway, _, err := d.client.Gateways.Get(ctx, data.ID.ValueString())
//	    if err != nil {
//	        resp.Diagnostics.AddError("Failed to Read Gateway", err.Error())
//	        return
//	    }
//
//	    // Map to model
//	    data.Name = types.StringValue(gateway.Name)
//	    data.Description = types.StringPointerValue(gateway.Description)
//
//	    // Save to state
//	    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
//	}
//
// # Resource Implementation Pattern
//
// Resources manage the lifecycle of ContextForge objects (create, read, update, delete).
// Each resource follows this pattern:
//
//  1. Define a struct implementing resource.Resource and resource.ResourceWithConfigure
//  2. Define a model struct with tfsdk tags
//  3. Implement required methods: Metadata, Schema, Create, Read, Update, Delete, Configure
//  4. Optionally implement ImportState for terraform import support
//  5. Register in provider's Resources() method
//
// # Client Access Pattern
//
// Data sources and resources access the ContextForge client via type assertion:
//
//	func (d *gatewayDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
//	    if req.ProviderData == nil {
//	        return
//	    }
//
//	    client, ok := req.ProviderData.(*contextforge.Client)
//	    if !ok {
//	        resp.Diagnostics.AddError(
//	            "Unexpected Data Source Configure Type",
//	            fmt.Sprintf("Expected *contextforge.Client, got: %T", req.ProviderData),
//	        )
//	        return
//	    }
//
//	    d.client = client
//	}
//
// # Testing Organization
//
// All tests are co-located in this package to avoid import cycles:
//
//   - provider_test.go - Shared test infrastructure (testAccProtoV6ProviderFactories, testAccPreCheck)
//   - data_source_<name>_test.go - Acceptance tests for specific data sources
//   - resource_<name>_test.go - Acceptance tests for specific resources
//
// Test utilities are unexported (lowercase) as they're package-internal.
//
// # Acceptance Testing
//
// Acceptance tests use the terraform-plugin-testing framework and require TF_ACC=1:
//
//	func TestAccGatewayDataSource_basic(t *testing.T) {
//	    resource.Test(t, resource.TestCase{
//	        PreCheck:                 func() { testAccPreCheck(t) },
//	        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//	        Steps: []resource.TestStep{
//	            {
//	                Config: testAccGatewayDataSourceConfig(gatewayID),
//	                Check: resource.ComposeAggregateTestCheckFunc(
//	                    resource.TestCheckResourceAttr("data.contextforge_gateway.test", "id", gatewayID),
//	                    resource.TestCheckResourceAttr("data.contextforge_gateway.test", "name", "test-gateway"),
//	                ),
//	            },
//	        },
//	    })
//	}
//
// Integration tests require a running ContextForge gateway. Use make targets:
//
//	make integration-test-setup    # Start gateway and MCP time server
//	make integration-test          # Run tests with TF_ACC=1
//	make integration-test-teardown # Clean up
//	make integration-test-all      # Full lifecycle
//
// # File Organization
//
// Files in this package follow these naming conventions:
//
//   - provider.go - Core provider implementation
//   - data_source_<name>.go - Data source implementation
//   - resource_<name>.go - Resource implementation
//   - data_source_<name>_test.go - Data source acceptance tests
//   - resource_<name>_test.go - Resource acceptance tests
//   - provider_test.go - Shared test utilities
//
// # Error Handling
//
// Use resp.Diagnostics for all errors and warnings:
//
//	resp.Diagnostics.AddError("Summary", "Detailed error message")
//	resp.Diagnostics.AddWarning("Summary", "Warning message")
//
// Always check resp.Diagnostics.HasError() before continuing after operations
// that may produce diagnostics.
//
// # Adding New Data Sources
//
// To add a new data source:
//
//  1. Create data_source_<name>.go with implementation
//  2. Create data_source_<name>_test.go with acceptance tests
//  3. Add NewXDataSource factory to DataSources() in provider.go
//  4. Follow naming convention: contextforge_<resource_type>
//  5. Document in CLAUDE.md and README.md
//
// # Adding New Resources
//
// To add a new resource:
//
//  1. Create resource_<name>.go with implementation
//  2. Create resource_<name>_test.go with acceptance tests
//  3. Add NewXResource factory to Resources() in provider.go
//  4. Implement ImportState for terraform import support
//  5. Follow naming convention: contextforge_<resource_type>
//  6. Document in CLAUDE.md and README.md
package provider
