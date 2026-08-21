package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type uptimeModel struct {
	OneHour         types.Float64 `tfsdk:"one_hour"`
	TwentyFourHours types.Float64 `tfsdk:"twenty_four_hours"`
	SevenDays       types.Float64 `tfsdk:"seven_days"`
	ThirtyDays      types.Float64 `tfsdk:"thirty_days"`
}

type monitorModel struct {
	ID                  types.String    `tfsdk:"id"`
	Name                types.String    `tfsdk:"name"`
	Description         types.String    `tfsdk:"description"`
	Tags                types.List      `tfsdk:"tags"`
	Group               types.String    `tfsdk:"group"`
	Type                types.String    `tfsdk:"type"`
	Enabled             types.Bool      `tfsdk:"enabled"`
	IntervalSeconds     types.Int64     `tfsdk:"interval_seconds"`
	TimeoutSeconds      types.Int64     `tfsdk:"timeout_seconds"`
	IPFamily            types.String    `tfsdk:"ip_family"`
	Target              types.String    `tfsdk:"target"`
	Method              types.String    `tfsdk:"method"`
	RequestHeaders      []keyValueModel `tfsdk:"request_headers"`
	RequestBody         types.String    `tfsdk:"request_body"`
	DNSQueryName        types.String    `tfsdk:"dns_query_name"`
	DNSQueryType        types.String    `tfsdk:"dns_query_type"`
	FollowRedirects     types.Bool      `tfsdk:"follow_redirects"`
	VerifyTLS           types.Bool      `tfsdk:"verify_tls"`
	ProxyURL            types.String    `tfsdk:"proxy_url"`
	RetentionDays       types.Int64     `tfsdk:"retention_days"`
	Conditions          types.List      `tfsdk:"conditions"`
	ProbeIDs            types.List      `tfsdk:"probe_ids"`
	ChannelIDs          types.List      `tfsdk:"channel_ids"`
	Status              types.String    `tfsdk:"status"`
	HeartbeatToken      types.String    `tfsdk:"heartbeat_token"`
	HeartbeatURL        types.String    `tfsdk:"heartbeat_url"`
	HeartbeatStartURL   types.String    `tfsdk:"heartbeat_start_url"`
	HeartbeatFinishURL  types.String    `tfsdk:"heartbeat_finish_url"`
	HeartbeatErrorURL   types.String    `tfsdk:"heartbeat_error_url"`
	StatusBadgeURL      types.String    `tfsdk:"status_badge_url"`
	StatusBadgeJSONURL  types.String    `tfsdk:"status_badge_json_url"`
	UptimeBadgeURL      types.String    `tfsdk:"uptime_badge_url"`
	UptimeBadgeJSONURL  types.String    `tfsdk:"uptime_badge_json_url"`
	LatencyBadgeURL     types.String    `tfsdk:"latency_badge_url"`
	LatencyBadgeJSONURL types.String    `tfsdk:"latency_badge_json_url"`
	BadgeMarkdown       types.String    `tfsdk:"badge_markdown"`
	Uptime              *uptimeModel    `tfsdk:"uptime"`
}

func (r *monitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedString := schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A Nominal monitor. Types: Http, GraphQL, Ping, Tcp, Dns, Tls, Heartbeat, Udp, WebSocket, Mysql, Redis, Postgres.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				Optional:           true,
				DeprecationMessage: "Removed from the Nominal API. Set tags instead. If tags are omitted, group is sent as a single tag.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Http, GraphQL, Ping, Tcp, Dns, Tls, Heartbeat, Udp, WebSocket, Mysql, Redis, or Postgres.",
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"interval_seconds": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				MarkdownDescription: "Check interval. Monitors with a `[DOMAIN_EXPIRATION]` condition must use at least 300 seconds.",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(10),
			},
			"ip_family": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Any"),
				MarkdownDescription: "Ipv4, Ipv6, or Any.",
			},
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL, host, or connection URL. Database monitors take `mysql://`, `postgres://`, or `redis://` URLs.",
			},
			"method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "HTTP method: GET, POST, PUT, PATCH, DELETE, or HEAD. GraphQL monitors default to POST.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP body, GraphQL query, UDP/WebSocket payload, custom SQL, or Redis command.",
			},
			"dns_query_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name to resolve for Dns monitors.",
			},
			"dns_query_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A, AAAA, CNAME, MX, NS, PTR, SRV, or TXT. Dns monitors default to A.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"proxy_url": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "HTTP or SOCKS proxy URL for Http, GraphQL, Tcp, Tls, WebSocket, and Redis monitors.",
			},
			"retention_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
			},
			"conditions": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Gatus-style expressions such as `[STATUS] == 200`. Omitted conditions use the type defaults.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"probe_ids": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Probe IDs to run this monitor. Omitted IDs attach Nominal's default probes.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_ids": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Notification channel IDs. Synced via `syncMonitorChannels`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"status": computedString,
			"heartbeat_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"heartbeat_url":          computedString,
			"heartbeat_start_url":    computedString,
			"heartbeat_finish_url":   computedString,
			"heartbeat_error_url":    computedString,
			"status_badge_url":       computedString,
			"status_badge_json_url":  computedString,
			"uptime_badge_url":       computedString,
			"uptime_badge_json_url":  computedString,
			"latency_badge_url":      computedString,
			"latency_badge_json_url": computedString,
			"badge_markdown":         computedString,
			"uptime": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Rolling uptime percentages from check results.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"one_hour":          schema.Float64Attribute{Computed: true},
					"twenty_four_hours": schema.Float64Attribute{Computed: true},
					"seven_days":        schema.Float64Attribute{Computed: true},
					"thirty_days":       schema.Float64Attribute{Computed: true},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"request_headers": schema.ListNestedBlock{
				MarkdownDescription: "HTTP or WebSocket request headers.",
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
		CreateMonitor gqlMonitor `json:"createMonitor"`
	}

	if err := r.client.Query(ctx, `
		mutation ($input: CreateMonitorInput!) {
			createMonitor(input: $input) {`+monitorSelection+`}
		}
	`, map[string]any{"input": r.input(ctx, plan)}, &out); err != nil {
		resp.Diagnostics.AddError("Create monitor failed", err.Error())
		return
	}

	plan.ID = types.StringValue(out.CreateMonitor.ID)
	if err := r.syncChannels(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Sync monitor channels failed", err.Error())
		return
	}

	refreshed, err := r.readMonitor(ctx, plan.ID.ValueString(), plan)
	if err != nil {
		resp.Diagnostics.AddError("Read monitor failed", err.Error())
		return
	}

	if refreshed == nil {
		state := monitorFromAPI(out.CreateMonitor, plan)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *monitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state monitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.readMonitor(ctx, state.ID.ValueString(), state)
	if err != nil {
		resp.Diagnostics.AddError("Read monitor failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan monitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!, $input: UpdateMonitorInput!) {
			updateMonitor(id: $id, input: $input) { id }
		}
	`, map[string]any{"id": plan.ID.ValueString(), "input": r.input(ctx, plan)}, nil); err != nil {
		resp.Diagnostics.AddError("Update monitor failed", err.Error())
		return
	}

	if err := r.syncChannels(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Sync monitor channels failed", err.Error())
		return
	}

	refreshed, err := r.readMonitor(ctx, plan.ID.ValueString(), plan)
	if err != nil {
		resp.Diagnostics.AddError("Read monitor failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state monitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!) { deleteMonitor(id: $id) }
	`, map[string]any{"id": state.ID.ValueString()}, nil); err != nil {
		resp.Diagnostics.AddError("Delete monitor failed", err.Error())
	}
}

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *monitorResource) input(ctx context.Context, model monitorModel) map[string]any {
	input := map[string]any{
		"name":            model.Name.ValueString(),
		"type":            model.Type.ValueString(),
		"target":          model.Target.ValueString(),
		"enabled":         model.Enabled.ValueBool(),
		"intervalSeconds": model.IntervalSeconds.ValueInt64(),
		"timeoutSeconds":  model.TimeoutSeconds.ValueInt64(),
		"ipFamily":        model.IPFamily.ValueString(),
		"followRedirects": model.FollowRedirects.ValueBool(),
		"verifyTls":       model.VerifyTLS.ValueBool(),
		"retentionDays":   model.RetentionDays.ValueInt64(),
	}

	if description, ok := optionalString(model.Description); ok {
		input["description"] = description
	} else if !model.Description.IsUnknown() {
		input["description"] = nil
	}

	if tags, ok := stringList(ctx, model.Tags); ok {
		input["tags"] = tags
	} else if group, ok := optionalString(model.Group); ok {
		input["tags"] = []string{group}
	}

	if method, ok := optionalString(model.Method); ok && method != "" {
		input["method"] = method
	}

	if body, ok := optionalString(model.RequestBody); ok {
		input["requestBody"] = body
	}

	if name, ok := optionalString(model.DNSQueryName); ok {
		input["dnsQueryName"] = name
	}

	if queryType, ok := optionalString(model.DNSQueryType); ok && queryType != "" {
		input["dnsQueryType"] = queryType
	}

	if proxyURL, ok := optionalString(model.ProxyURL); ok {
		input["proxyUrl"] = proxyURL
	} else if !model.ProxyURL.IsUnknown() {
		input["proxyUrl"] = nil
	}

	headers := keyValuesInput(model.RequestHeaders)
	if headers == nil {
		headers = []map[string]string{}
	}
	input["requestHeaders"] = headers

	if conditions, ok := stringList(ctx, model.Conditions); ok {
		input["conditions"] = conditions
	}

	if probeIDs, ok := stringList(ctx, model.ProbeIDs); ok {
		input["probeIds"] = probeIDs
	}

	return input
}

func (r *monitorResource) syncChannels(ctx context.Context, model monitorModel) error {
	channelIDs, ok := stringList(ctx, model.ChannelIDs)
	if !ok {
		return nil
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

func (r *monitorResource) readMonitor(ctx context.Context, id string, previous monitorModel) (*monitorModel, error) {
	var out struct {
		Monitor *gqlMonitor `json:"monitor"`
	}

	if err := r.client.Query(ctx, `
		query ($id: ID!) {
			monitor(id: $id) {`+monitorSelection+`}
		}
	`, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}

	if out.Monitor == nil {
		return nil, nil
	}

	state := monitorFromAPI(*out.Monitor, previous)
	return &state, nil
}

func monitorFromAPI(monitor gqlMonitor, previous monitorModel) monitorModel {
	state := monitorModel{
		ID:                  types.StringValue(monitor.ID),
		Name:                types.StringValue(monitor.Name),
		Description:         stringOrNull(monitor.Description),
		Tags:                stringListValue(monitor.Tags),
		Group:               previous.Group,
		Type:                types.StringValue(monitor.Type),
		Enabled:             types.BoolValue(monitor.Enabled),
		IntervalSeconds:     types.Int64Value(monitor.IntervalSeconds),
		TimeoutSeconds:      types.Int64Value(monitor.TimeoutSeconds),
		IPFamily:            types.StringValue(monitor.IPFamily),
		Target:              types.StringValue(monitor.Target),
		Method:              stringOrNull(monitor.Method),
		RequestHeaders:      keyValuesModel(monitor.RequestHeaders),
		RequestBody:         stringOrNull(monitor.RequestBody),
		DNSQueryName:        stringOrNull(monitor.DNSQueryName),
		DNSQueryType:        stringOrNull(monitor.DNSQueryType),
		FollowRedirects:     types.BoolValue(monitor.FollowRedirects),
		VerifyTLS:           types.BoolValue(monitor.VerifyTLS),
		ProxyURL:            stringOrNull(monitor.ProxyURL),
		RetentionDays:       types.Int64Value(monitor.RetentionDays),
		Conditions:          stringListValue(expressionsFromConditions(monitor.Conditions)),
		ProbeIDs:            stringListValue(idsFromNodes(monitor.Probes)),
		ChannelIDs:          stringListValue(idsFromNodes(monitor.NotificationChannels)),
		Status:              types.StringValue(monitor.Status),
		HeartbeatToken:      stringOrNull(monitor.HeartbeatToken),
		HeartbeatURL:        stringOrNull(monitor.HeartbeatURL),
		HeartbeatStartURL:   stringOrNull(monitor.HeartbeatStartURL),
		HeartbeatFinishURL:  stringOrNull(monitor.HeartbeatFinishURL),
		HeartbeatErrorURL:   stringOrNull(monitor.HeartbeatErrorURL),
		StatusBadgeURL:      stringValueOrNull(monitor.StatusBadgeURL),
		StatusBadgeJSONURL:  stringValueOrNull(monitor.StatusBadgeJSONURL),
		UptimeBadgeURL:      stringValueOrNull(monitor.UptimeBadgeURL),
		UptimeBadgeJSONURL:  stringValueOrNull(monitor.UptimeBadgeJSONURL),
		LatencyBadgeURL:     stringValueOrNull(monitor.LatencyBadgeURL),
		LatencyBadgeJSONURL: stringValueOrNull(monitor.LatencyBadgeJSONURL),
		BadgeMarkdown:       stringValueOrNull(monitor.BadgeMarkdown),
		Uptime: &uptimeModel{
			OneHour:         floatOrNull(monitor.Uptime.OneHour),
			TwentyFourHours: floatOrNull(monitor.Uptime.TwentyFourHours),
			SevenDays:       floatOrNull(monitor.Uptime.SevenDays),
			ThirtyDays:      floatOrNull(monitor.Uptime.ThirtyDays),
		},
	}

	if previous.ChannelIDs.IsNull() && len(monitor.NotificationChannels) == 0 {
		state.ChannelIDs = types.ListNull(types.StringType)
	}

	return state
}
