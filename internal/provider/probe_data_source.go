package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ datasource.DataSource = (*probeDataSource)(nil)

func NewProbeDataSource() datasource.DataSource {
	return &probeDataSource{}
}

type probeDataSource struct {
	client *client.Client
}

type probeDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Slug      types.String `tfsdk:"slug"`
	Name      types.String `tfsdk:"name"`
	Queue     types.String `tfsdk:"queue"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	IsDefault types.Bool   `tfsdk:"is_default"`
}

func (d *probeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_probe"
}

func (d *probeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a Nominal probe by slug or id. Probes are created by Nominal itself; this data source is read-only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Probe slug such as `local` or `us-east`.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"queue": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"is_default": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "When true, new monitors attach this probe unless probe_ids is set.",
			},
		},
	}
}

func (d *probeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}

	d.client = api
}

func (d *probeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config probeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, hasID := optionalString(config.ID)
	slug, hasSlug := optionalString(config.Slug)
	if !hasID && !hasSlug {
		resp.Diagnostics.AddError("Missing probe lookup", "Set slug or id.")
		return
	}

	probes, err := listProbes(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Read probes failed", err.Error())
		return
	}

	var match *gqlProbe
	for i := range probes {
		probe := probes[i]
		if hasID && probe.ID == id {
			match = &probe
			break
		}
		if hasSlug && probe.Slug == slug {
			match = &probe
			break
		}
	}

	if match == nil {
		resp.Diagnostics.AddError("Probe not found", "No probe matched the given slug or id.")
		return
	}

	state := probeDataSourceModel{
		ID:        types.StringValue(match.ID),
		Slug:      types.StringValue(match.Slug),
		Name:      types.StringValue(match.Name),
		Queue:     types.StringValue(match.Queue),
		Enabled:   types.BoolValue(match.Enabled),
		IsDefault: types.BoolValue(match.IsDefault),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
