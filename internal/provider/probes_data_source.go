package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ datasource.DataSource = (*probesDataSource)(nil)

func NewProbesDataSource() datasource.DataSource {
	return &probesDataSource{}
}

type probesDataSource struct {
	client *client.Client
}

type probesDataSourceModel struct {
	Probes []probeDataSourceModel `tfsdk:"probes"`
}

func (d *probesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_probes"
}

func (d *probesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "All Nominal probes.",
		Attributes: map[string]schema.Attribute{
			"probes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"slug":       schema.StringAttribute{Computed: true},
						"name":       schema.StringAttribute{Computed: true},
						"queue":      schema.StringAttribute{Computed: true},
						"enabled":    schema.BoolAttribute{Computed: true},
						"is_default": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *probesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *probesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	probes, err := listProbes(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Read probes failed", err.Error())
		return
	}

	state := probesDataSourceModel{Probes: make([]probeDataSourceModel, 0, len(probes))}
	for _, probe := range probes {
		state.Probes = append(state.Probes, probeDataSourceModel{
			ID:        types.StringValue(probe.ID),
			Slug:      types.StringValue(probe.Slug),
			Name:      types.StringValue(probe.Name),
			Queue:     types.StringValue(probe.Queue),
			Enabled:   types.BoolValue(probe.Enabled),
			IsDefault: types.BoolValue(probe.IsDefault),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func listProbes(ctx context.Context, api *client.Client) ([]gqlProbe, error) {
	var out struct {
		Probes []gqlProbe `json:"probes"`
	}

	if err := api.Query(ctx, `
		query {
			probes { id slug name queue enabled is_default }
		}
	`, nil, &out); err != nil {
		return nil, err
	}

	return out.Probes, nil
}
