package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ resource.Resource = (*statusPageResource)(nil)
var _ resource.ResourceWithImportState = (*statusPageResource)(nil)

func NewStatusPageResource() resource.Resource {
	return &statusPageResource{}
}

type statusPageResource struct {
	client *client.Client
}

type statusPageMonitorModel struct {
	MonitorID  types.String `tfsdk:"monitor_id"`
	PublicName types.String `tfsdk:"public_name"`
}

type statusPageModel struct {
	ID                types.String             `tfsdk:"id"`
	Name              types.String             `tfsdk:"name"`
	Slug              types.String             `tfsdk:"slug"`
	CustomDomain      types.String             `tfsdk:"custom_domain"`
	Headline          types.String             `tfsdk:"headline"`
	Description       types.String             `tfsdk:"description"`
	LogoURL           types.String             `tfsdk:"logo_url"`
	FaviconURL        types.String             `tfsdk:"favicon_url"`
	FooterText        types.String             `tfsdk:"footer_text"`
	CustomCSS         types.String             `tfsdk:"custom_css"`
	Theme             types.String             `tfsdk:"theme"`
	Published         types.Bool               `tfsdk:"published"`
	ShowTargets       types.Bool               `tfsdk:"show_targets"`
	Password          types.String             `tfsdk:"password"`
	PasswordProtected types.Bool               `tfsdk:"password_protected"`
	RefreshSeconds    types.Int64              `tfsdk:"refresh_seconds"`
	Monitors          []statusPageMonitorModel `tfsdk:"monitor"`
	PathURL           types.String             `tfsdk:"path_url"`
	PublicURL         types.String             `tfsdk:"public_url"`
	Health            types.String             `tfsdk:"health"`
}

const statusPageSelection = `
	id
	name
	slug
	custom_domain
	headline
	description
	logo_url
	favicon_url
	footer_text
	custom_css
	theme
	published
	show_targets
	passwordProtected
	refresh_seconds
	pathUrl
	publicUrl
	health
	listings {
		public_name
		monitor { id }
	}
`

func (r *statusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedString := schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A public Nominal status page, including listed monitors, branding, and an optional password.",
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
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path slug served at `/status/{slug}`.",
			},
			"custom_domain": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Hostname for the page at `/`. Scheme prefixes are stripped by the API.",
			},
			"headline": schema.StringAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"logo_url": schema.StringAttribute{
				Optional: true,
			},
			"favicon_url": schema.StringAttribute{
				Optional: true,
			},
			"footer_text": schema.StringAttribute{
				Optional: true,
			},
			"custom_css": schema.StringAttribute{
				Optional: true,
			},
			"theme": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Dark"),
				MarkdownDescription: "Dark or Light.",
			},
			"published": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"show_targets": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When true, the public page shows monitor targets.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Visitor password. Never returned by GraphQL; omit to leave the current password unchanged on update, set `\"\"` to clear it.",
			},
			"password_protected": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"refresh_seconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
			},
			"path_url":   computedString,
			"public_url": computedString,
			"health":     computedString,
		},
		Blocks: map[string]schema.Block{
			"monitor": schema.ListNestedBlock{
				MarkdownDescription: "Monitors listed on the page, in order.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"monitor_id": schema.StringAttribute{
							Required: true,
						},
						"public_name": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Public label. Targets stay hidden unless show_targets is true.",
						},
					},
				},
			},
		},
	}
}

func (r *statusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan statusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		CreateStatusPage gqlStatusPage `json:"createStatusPage"`
	}

	if err := r.client.Query(ctx, `
		mutation ($input: CreateStatusPageInput!) {
			createStatusPage(input: $input) {`+statusPageSelection+`}
		}
	`, map[string]any{"input": r.input(plan)}, &out); err != nil {
		resp.Diagnostics.AddError("Create status page failed", err.Error())
		return
	}

	state := statusPageFromAPI(out.CreateStatusPage, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state statusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.readStatusPage(ctx, state.ID.ValueString(), state)
	if err != nil {
		resp.Diagnostics.AddError("Read status page failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *statusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan statusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!, $input: UpdateStatusPageInput!) {
			updateStatusPage(id: $id, input: $input) { id }
		}
	`, map[string]any{"id": plan.ID.ValueString(), "input": r.input(plan)}, nil); err != nil {
		resp.Diagnostics.AddError("Update status page failed", err.Error())
		return
	}

	refreshed, err := r.readStatusPage(ctx, plan.ID.ValueString(), plan)
	if err != nil {
		resp.Diagnostics.AddError("Read status page failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *statusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state statusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!) { deleteStatusPage(id: $id) }
	`, map[string]any{"id": state.ID.ValueString()}, nil); err != nil {
		resp.Diagnostics.AddError("Delete status page failed", err.Error())
	}
}

func (r *statusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *statusPageResource) input(model statusPageModel) map[string]any {
	input := map[string]any{
		"name":           model.Name.ValueString(),
		"slug":           model.Slug.ValueString(),
		"theme":          model.Theme.ValueString(),
		"published":      model.Published.ValueBool(),
		"showTargets":    model.ShowTargets.ValueBool(),
		"refreshSeconds": model.RefreshSeconds.ValueInt64(),
	}

	if value, ok := optionalString(model.CustomDomain); ok {
		input["customDomain"] = value
	} else if !model.CustomDomain.IsUnknown() {
		input["customDomain"] = nil
	}

	if value, ok := optionalString(model.Headline); ok {
		input["headline"] = value
	}

	if value, ok := optionalString(model.Description); ok {
		input["description"] = value
	}

	if value, ok := optionalString(model.LogoURL); ok {
		input["logoUrl"] = value
	}

	if value, ok := optionalString(model.FaviconURL); ok {
		input["faviconUrl"] = value
	}

	if value, ok := optionalString(model.FooterText); ok {
		input["footerText"] = value
	}

	if value, ok := optionalString(model.CustomCSS); ok {
		input["customCss"] = value
	}

	if value, ok := optionalString(model.Password); ok {
		input["password"] = value
	}

	monitors := make([]map[string]any, 0, len(model.Monitors))
	for _, item := range model.Monitors {
		entry := map[string]any{
			"monitorId": item.MonitorID.ValueString(),
		}
		if name, ok := optionalString(item.PublicName); ok {
			entry["publicName"] = name
		}
		monitors = append(monitors, entry)
	}
	input["monitors"] = monitors

	return input
}

func (r *statusPageResource) readStatusPage(ctx context.Context, id string, previous statusPageModel) (*statusPageModel, error) {
	var out struct {
		StatusPage *gqlStatusPage `json:"statusPage"`
	}

	if err := r.client.Query(ctx, `
		query ($id: ID!) {
			statusPage(id: $id) {`+statusPageSelection+`}
		}
	`, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}

	if out.StatusPage == nil {
		return nil, nil
	}

	state := statusPageFromAPI(*out.StatusPage, previous)
	return &state, nil
}

func statusPageFromAPI(page gqlStatusPage, previous statusPageModel) statusPageModel {
	monitors := make([]statusPageMonitorModel, 0, len(page.Listings))
	for _, listing := range page.Listings {
		monitors = append(monitors, statusPageMonitorModel{
			MonitorID:  types.StringValue(listing.Monitor.ID),
			PublicName: stringOrNull(listing.PublicName),
		})
	}

	return statusPageModel{
		ID:                types.StringValue(page.ID),
		Name:              types.StringValue(page.Name),
		Slug:              types.StringValue(page.Slug),
		CustomDomain:      stringOrNull(page.CustomDomain),
		Headline:          stringOrNull(page.Headline),
		Description:       stringOrNull(page.Description),
		LogoURL:           stringOrNull(page.LogoURL),
		FaviconURL:        stringOrNull(page.FaviconURL),
		FooterText:        stringOrNull(page.FooterText),
		CustomCSS:         stringOrNull(page.CustomCSS),
		Theme:             types.StringValue(page.Theme),
		Published:         types.BoolValue(page.Published),
		ShowTargets:       types.BoolValue(page.ShowTargets),
		Password:          previous.Password,
		PasswordProtected: types.BoolValue(page.PasswordProtected),
		RefreshSeconds:    types.Int64Value(page.RefreshSeconds),
		Monitors:          monitors,
		PathURL:           types.StringValue(page.PathURL),
		PublicURL:         types.StringValue(page.PublicURL),
		Health:            types.StringValue(page.Health),
	}
}
