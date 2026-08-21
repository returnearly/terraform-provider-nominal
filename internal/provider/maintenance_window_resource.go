package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ resource.Resource = (*maintenanceWindowResource)(nil)
var _ resource.ResourceWithImportState = (*maintenanceWindowResource)(nil)

func NewMaintenanceWindowResource() resource.Resource {
	return &maintenanceWindowResource{}
}

type maintenanceWindowResource struct {
	client *client.Client
}

type maintenanceWindowModel struct {
	ID           types.String `tfsdk:"id"`
	Title        types.String `tfsdk:"title"`
	Message      types.String `tfsdk:"message"`
	StartsAt     types.String `tfsdk:"starts_at"`
	EndsAt       types.String `tfsdk:"ends_at"`
	AppliesToAll types.Bool   `tfsdk:"applies_to_all"`
	MonitorIDs   types.List   `tfsdk:"monitor_ids"`
	Phase        types.String `tfsdk:"phase"`
}

const maintenanceWindowSelection = `
	id
	title
	message
	starts_at
	ends_at
	applies_to_all
	phase
	monitors { id }
`

func (r *maintenanceWindowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

func (r *maintenanceWindowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A maintenance window that suppresses alerts. Dates use Nominal's GraphQL DateTime format `YYYY-MM-DD HH:MM:SS`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required: true,
			},
			"message": schema.StringAttribute{
				Optional: true,
			},
			"starts_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Window start. Defaults to now when omitted. Format: `YYYY-MM-DD HH:MM:SS`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ends_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Window end. Omit for an open-ended window. Format: `YYYY-MM-DD HH:MM:SS`.",
			},
			"applies_to_all": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When true, the window covers every monitor. Otherwise set monitor_ids.",
			},
			"monitor_ids": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Monitors covered by this window. Required unless applies_to_all is true.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"phase": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "scheduled, active, or ended.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *maintenanceWindowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *maintenanceWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan maintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		CreateMaintenanceWindow gqlMaintenanceWindow `json:"createMaintenanceWindow"`
	}

	if err := r.client.Query(ctx, `
		mutation ($input: CreateMaintenanceWindowInput!) {
			createMaintenanceWindow(input: $input) {`+maintenanceWindowSelection+`}
		}
	`, map[string]any{"input": r.input(ctx, plan)}, &out); err != nil {
		resp.Diagnostics.AddError("Create maintenance window failed", err.Error())
		return
	}

	state := maintenanceWindowFromAPI(out.CreateMaintenanceWindow)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *maintenanceWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state maintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.readWindow(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read maintenance window failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *maintenanceWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan maintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!, $input: UpdateMaintenanceWindowInput!) {
			updateMaintenanceWindow(id: $id, input: $input) { id }
		}
	`, map[string]any{"id": plan.ID.ValueString(), "input": r.input(ctx, plan)}, nil); err != nil {
		resp.Diagnostics.AddError("Update maintenance window failed", err.Error())
		return
	}

	refreshed, err := r.readWindow(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read maintenance window failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *maintenanceWindowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state maintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!) { deleteMaintenanceWindow(id: $id) }
	`, map[string]any{"id": state.ID.ValueString()}, nil); err != nil {
		resp.Diagnostics.AddError("Delete maintenance window failed", err.Error())
	}
}

func (r *maintenanceWindowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *maintenanceWindowResource) input(ctx context.Context, model maintenanceWindowModel) map[string]any {
	input := map[string]any{
		"title":        model.Title.ValueString(),
		"appliesToAll": model.AppliesToAll.ValueBool(),
	}

	if message, ok := optionalString(model.Message); ok {
		input["message"] = message
	} else if !model.Message.IsUnknown() {
		input["message"] = nil
	}

	if startsAt, ok := optionalString(model.StartsAt); ok && startsAt != "" {
		input["startsAt"] = startsAt
	}

	if endsAt, ok := optionalString(model.EndsAt); ok {
		input["endsAt"] = endsAt
	} else if !model.EndsAt.IsUnknown() {
		input["endsAt"] = nil
	}

	if monitorIDs, ok := stringList(ctx, model.MonitorIDs); ok {
		input["monitorIds"] = monitorIDs
	}

	return input
}

func (r *maintenanceWindowResource) readWindow(ctx context.Context, id string) (*maintenanceWindowModel, error) {
	var out struct {
		MaintenanceWindow *gqlMaintenanceWindow `json:"maintenanceWindow"`
	}

	if err := r.client.Query(ctx, `
		query ($id: ID!) {
			maintenanceWindow(id: $id) {`+maintenanceWindowSelection+`}
		}
	`, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}

	if out.MaintenanceWindow == nil {
		return nil, nil
	}

	state := maintenanceWindowFromAPI(*out.MaintenanceWindow)
	return &state, nil
}

func maintenanceWindowFromAPI(window gqlMaintenanceWindow) maintenanceWindowModel {
	return maintenanceWindowModel{
		ID:           types.StringValue(window.ID),
		Title:        types.StringValue(window.Title),
		Message:      stringOrNull(window.Message),
		StartsAt:     types.StringValue(window.StartsAt),
		EndsAt:       stringOrNull(window.EndsAt),
		AppliesToAll: types.BoolValue(window.AppliesToAll),
		MonitorIDs:   stringListValue(idsFromNodes(window.Monitors)),
		Phase:        types.StringValue(window.Phase),
	}
}
