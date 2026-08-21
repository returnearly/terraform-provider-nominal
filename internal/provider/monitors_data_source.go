package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ datasource.DataSource = (*monitorsDataSource)(nil)

func NewMonitorsDataSource() datasource.DataSource {
	return &monitorsDataSource{}
}

type monitorsDataSource struct {
	client *client.Client
}

type monitorSummaryModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Status  types.String `tfsdk:"status"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Target  types.String `tfsdk:"target"`
	Tags    types.List   `tfsdk:"tags"`
}

type monitorsDataSourceModel struct {
	Tag      types.String          `tfsdk:"tag"`
	Monitors []monitorSummaryModel `tfsdk:"monitors"`
}

func (d *monitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *monitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List Nominal monitors, optionally filtered by tag.",
		Attributes: map[string]schema.Attribute{
			"tag": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "When set, only monitors with this tag are returned.",
			},
			"monitors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"type":    schema.StringAttribute{Computed: true},
						"status":  schema.StringAttribute{Computed: true},
						"enabled": schema.BoolAttribute{Computed: true},
						"target":  schema.StringAttribute{Computed: true},
						"tags": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *monitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *monitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config monitorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		Monitors []struct {
			ID      string   `json:"id"`
			Name    string   `json:"name"`
			Type    string   `json:"type"`
			Status  string   `json:"status"`
			Enabled bool     `json:"enabled"`
			Target  string   `json:"target"`
			Tags    []string `json:"tags"`
		} `json:"monitors"`
	}

	variables := map[string]any{}
	query := `
		query ($tag: String) {
			monitors(tag: $tag) {
				id name type status enabled target tags
			}
		}
	`
	if tag, ok := optionalString(config.Tag); ok {
		variables["tag"] = tag
	} else {
		variables["tag"] = nil
	}

	if err := d.client.Query(ctx, query, variables, &out); err != nil {
		resp.Diagnostics.AddError("Read monitors failed", err.Error())
		return
	}

	state := monitorsDataSourceModel{
		Tag:      config.Tag,
		Monitors: make([]monitorSummaryModel, 0, len(out.Monitors)),
	}
	for _, monitor := range out.Monitors {
		state.Monitors = append(state.Monitors, monitorSummaryModel{
			ID:      types.StringValue(monitor.ID),
			Name:    types.StringValue(monitor.Name),
			Type:    types.StringValue(monitor.Type),
			Status:  types.StringValue(monitor.Status),
			Enabled: types.BoolValue(monitor.Enabled),
			Target:  types.StringValue(monitor.Target),
			Tags:    stringListValue(monitor.Tags),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
