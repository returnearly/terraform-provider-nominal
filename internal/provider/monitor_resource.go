package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ resource.Resource = (*monitorResource)(nil)
var _ resource.ResourceWithImportState = (*monitorResource)(nil)

func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

type monitorResource struct {
	client *client.Client
}

type keyValueModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type monitorModel struct {
	ID               types.String    `tfsdk:"id"`
	Name             types.String    `tfsdk:"name"`
	Group            types.String    `tfsdk:"group"`
	Type             types.String    `tfsdk:"type"`
	Enabled          types.Bool      `tfsdk:"enabled"`
	IntervalSeconds  types.Int64     `tfsdk:"interval_seconds"`
	TimeoutSeconds   types.Int64     `tfsdk:"timeout_seconds"`
	IPFamily         types.String    `tfsdk:"ip_family"`
	Target           types.String    `tfsdk:"target"`
	Method           types.String    `tfsdk:"method"`
	RequestHeaders   []keyValueModel `tfsdk:"request_headers"`
	RequestBody      types.String    `tfsdk:"request_body"`
	FollowRedirects  types.Bool      `tfsdk:"follow_redirects"`
	VerifyTLS        types.Bool      `tfsdk:"verify_tls"`
	RetentionDays    types.Int64     `tfsdk:"retention_days"`
	Conditions       types.List      `tfsdk:"conditions"`
	ProbeIDs         types.List      `tfsdk:"probe_ids"`
	ChannelIDs       types.List      `tfsdk:"channel_ids"`
}

func (r *monitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Nominal HTTP or ping monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"group": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Http or Ping",
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"interval_seconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(10),
			},
			"ip_family": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Any"),
			},
			"target": schema.StringAttribute{
				Required: true,
			},
			"method": schema.StringAttribute{
				Optional: true,
			},
			"request_body": schema.StringAttribute{
				Optional: true,
			},
			"follow_redirects": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"verify_tls": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"retention_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
			},
			"conditions": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"probe_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"channel_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"request_headers": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{Required: true},
						"value": schema.StringAttribute{Required: true, Sensitive: true},
					},
				},
			},
		},
	}
}

func (r *monitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}

	r.client = api
}

func (r *monitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan monitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		CreateMonitor map[string]any `json:"createMonitor"`
	}

	if err := r.client.Query(ctx, `
		mutation ($input: CreateMonitorInput!) {
			createMonitor(input: $input) { id }
		}
	`, map[string]any{"input": r.input(ctx, plan)}, &out); err != nil {
		resp.Diagnostics.AddError("Create monitor failed", err.Error())
		return
	}

	id, _ := out.CreateMonitor["id"].(string)
	plan.ID = types.StringValue(id)

	if err := r.syncChannels(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Sync monitor channels failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *monitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state monitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		Monitor *struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Target string `json:"target"`
		} `json:"monitor"`
	}

	if err := r.client.Query(ctx, `
		query ($id: ID!) {
			monitor(id: $id) { id name target }
		}
	`, map[string]any{"id": state.ID.ValueString()}, &out); err != nil {
		resp.Diagnostics.AddError("Read monitor failed", err.Error())
		return
	}

	if out.Monitor == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan monitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		UpdateMonitor map[string]any `json:"updateMonitor"`
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!, $input: UpdateMonitorInput!) {
			updateMonitor(id: $id, input: $input) { id }
		}
	`, map[string]any{"id": plan.ID.ValueString(), "input": r.input(ctx, plan)}, &out); err != nil {
		resp.Diagnostics.AddError("Update monitor failed", err.Error())
		return
	}

	if err := r.syncChannels(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Sync monitor channels failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state monitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		DeleteMonitor bool `json:"deleteMonitor"`
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!) { deleteMonitor(id: $id) }
	`, map[string]any{"id": state.ID.ValueString()}, &out); err != nil {
		resp.Diagnostics.AddError("Delete monitor failed", err.Error())
	}
}

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *monitorResource) input(ctx context.Context, model monitorModel) map[string]any {
	input := map[string]any{
		"name":             model.Name.ValueString(),
		"type":             model.Type.ValueString(),
		"target":           model.Target.ValueString(),
		"enabled":          model.Enabled.ValueBool(),
		"intervalSeconds":  model.IntervalSeconds.ValueInt64(),
		"timeoutSeconds":   model.TimeoutSeconds.ValueInt64(),
		"ipFamily":         model.IPFamily.ValueString(),
		"followRedirects":  model.FollowRedirects.ValueBool(),
		"verifyTls":        model.VerifyTLS.ValueBool(),
		"retentionDays":    model.RetentionDays.ValueInt64(),
	}

	if !model.Group.IsNull() && !model.Group.IsUnknown() {
		input["group"] = model.Group.ValueString()
	}

	if !model.Method.IsNull() && !model.Method.IsUnknown() && model.Method.ValueString() != "" {
		input["method"] = model.Method.ValueString()
	}

	if !model.RequestBody.IsNull() && !model.RequestBody.IsUnknown() {
		input["requestBody"] = model.RequestBody.ValueString()
	}

	if len(model.RequestHeaders) > 0 {
		headers := make([]map[string]string, 0, len(model.RequestHeaders))
		for _, header := range model.RequestHeaders {
			headers = append(headers, map[string]string{
				"key":   header.Key.ValueString(),
				"value": header.Value.ValueString(),
			})
		}
		input["requestHeaders"] = headers
	}

	if !model.Conditions.IsNull() && !model.Conditions.IsUnknown() {
		var conditions []string
		_ = model.Conditions.ElementsAs(ctx, &conditions, false)
		input["conditions"] = conditions
	}

	if !model.ProbeIDs.IsNull() && !model.ProbeIDs.IsUnknown() {
		var probeIDs []string
		_ = model.ProbeIDs.ElementsAs(ctx, &probeIDs, false)
		input["probeIds"] = probeIDs
	}

	return input
}

func (r *monitorResource) syncChannels(ctx context.Context, model monitorModel) error {
	if model.ChannelIDs.IsNull() || model.ChannelIDs.IsUnknown() {
		return nil
	}

	var channelIDs []string
	_ = model.ChannelIDs.ElementsAs(ctx, &channelIDs, false)
	if channelIDs == nil {
		channelIDs = []string{}
	}

	return r.client.Query(ctx, `
		mutation ($monitorId: ID!, $channelIds: [ID!]!) {
			syncMonitorChannels(monitorId: $monitorId, channelIds: $channelIds) { id }
		}
	`, map[string]any{
		"monitorId":  model.ID.ValueString(),
		"channelIds": channelIDs,
	}, nil)
}
