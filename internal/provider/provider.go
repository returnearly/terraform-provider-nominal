package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ provider.Provider = (*nominalProvider)(nil)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &nominalProvider{version: version}
	}
}

type nominalProvider struct {
	version string
}

type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *nominalProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nominal"
	resp.Version = p.version
}

func (p *nominalProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Nominal monitors through GraphQL mutations. GraphQL `errors[]` are treated as failed applies even when HTTP status is 200.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "GraphQL endpoint, e.g. https://nominal.example.com/graphql",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Sanctum bearer token. Falls back to NOMINAL_TOKEN.",
			},
		},
	}
}

func (p *nominalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("NOMINAL_TOKEN")
	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		token = config.Token.ValueString()
	}

	if config.Endpoint.IsNull() || config.Endpoint.ValueString() == "" {
		resp.Diagnostics.AddError("Missing endpoint", "Set the GraphQL endpoint.")
		return
	}

	if token == "" {
		resp.Diagnostics.AddError("Missing token", "Set token or NOMINAL_TOKEN.")
		return
	}

	api := client.New(config.Endpoint.ValueString(), token)
	resp.ResourceData = api
	resp.DataSourceData = api
}

func (p *nominalProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMonitorResource,
		NewNotificationChannelResource,
	}
}

func (p *nominalProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
