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
	openai "github.com/sashabaranov/go-openai"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &openaiProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &openaiProvider{
			version: version,
		}
	}
}

// openaiProvider is the provider implementation
type openaiProvider struct {
	version string
}

// openaiProviderModel  maps provider schema data to a Go type
type openaiProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *openaiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openai"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *openaiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with OpenAI.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "The OpenAI API key for API operations. May also be provided via OPENAI_API_KEY environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *openaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring OpenAi client")

	// Retrieve provider data from configuration
	var config openaiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing OpenAI API key",
			"The provider cannot create the OpenAI API client as there is a missing or empty value for the OpenAI API key. "+
				"Set the api_key value in the configuration or use the OPENAI_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "api_key")

	tflog.Debug(ctx, "Creating OpenAI client")

	// Create a new OpenAI client using the configuration values
	client := openai.NewClient(apiKey)

	// Make the OpenAI client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured OpenAI client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *openaiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAssistantDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *openaiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAssistantResource,
		NewAssistantFileResource,
	}
}
