package provider

import (
	"context"
	"os"
	"sort"

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
	Headers  types.Map    `tfsdk:"headers"`
}

func (p *nominalProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nominal"
	resp.Version = p.version
}

func (p *nominalProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Nominal monitors, notification channels, status pages, and maintenance windows through GraphQL. GraphQL `errors[]` are treated as failed applies even when HTTP status is 200.",
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
			"headers": schema.MapAttribute{
				Optional:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extra HTTP headers on every GraphQL request. Use this for Cloudflare Access service tokens (`CF-Access-Client-Id`, `CF-Access-Client-Secret`) or similar proxies. `Authorization`, `Content-Type`, and `Accept` are reserved and ignored.",
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

	headers, reserved := providerHeaders(ctx, config.Headers)
	for _, name := range reserved {
		resp.Diagnostics.AddWarning(
			"Reserved HTTP header ignored",
			name+" is set by the provider and cannot be overridden in headers.",
		)
	}

	api := client.New(config.Endpoint.ValueString(), token).WithHeaders(headers)
	resp.ResourceData = api
	resp.DataSourceData = api
}

func providerHeaders(ctx context.Context, raw types.Map) (map[string]string, []string) {
	if raw.IsNull() || raw.IsUnknown() {
		return nil, nil
	}

	var values map[string]string
	_ = raw.ElementsAs(ctx, &values, false)

	var reserved []string
	for key := range values {
		if client.ReservedHeader(key) {
			reserved = append(reserved, key)
		}
	}
	sort.Strings(reserved)

	return client.SanitizeHeaders(values), reserved
}

func (p *nominalProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMonitorResource,
		NewNotificationChannelResource,
		NewStatusPageResource,
		NewMaintenanceWindowResource,
	}
}

func (p *nominalProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProbeDataSource,
		NewProbesDataSource,
		NewMonitorsDataSource,
	}
}
