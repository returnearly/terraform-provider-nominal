package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

var _ resource.Resource = (*notificationChannelResource)(nil)
var _ resource.ResourceWithImportState = (*notificationChannelResource)(nil)

func NewNotificationChannelResource() resource.Resource {
	return &notificationChannelResource{}
}

type notificationChannelResource struct {
	client *client.Client
}

type notificationChannelModel struct {
	ID     types.String    `tfsdk:"id"`
	Name   types.String    `tfsdk:"name"`
	Type   types.String    `tfsdk:"type"`
	Config []keyValueModel `tfsdk:"config"`
}

func (r *notificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *notificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Nominal notification channel (Mail, Slack, MicrosoftTeams, Discord, Webhook, or Pagerduty).",
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
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Mail, Slack, MicrosoftTeams, Discord, Webhook, or Pagerduty.",
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.ListNestedBlock{
				MarkdownDescription: "Channel settings as key/value pairs (`url`, `to`, `routing_key`, ...).",
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

func (r *notificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out struct {
		CreateNotificationChannel gqlNotificationChannel `json:"createNotificationChannel"`
	}

	if err := r.client.Query(ctx, `
		mutation ($input: CreateNotificationChannelInput!) {
			createNotificationChannel(input: $input) {
				id name type
				config { key value }
			}
		}
	`, map[string]any{"input": r.input(plan)}, &out); err != nil {
		resp.Diagnostics.AddError("Create notification channel failed", err.Error())
		return
	}

	state := notificationChannelFromAPI(out.CreateNotificationChannel)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.readChannel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read notification channel failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *notificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!, $input: UpdateNotificationChannelInput!) {
			updateNotificationChannel(id: $id, input: $input) { id }
		}
	`, map[string]any{"id": plan.ID.ValueString(), "input": r.input(plan)}, nil); err != nil {
		resp.Diagnostics.AddError("Update notification channel failed", err.Error())
		return
	}

	refreshed, err := r.readChannel(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read notification channel failed", err.Error())
		return
	}

	if refreshed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, refreshed)...)
}

func (r *notificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!) { deleteNotificationChannel(id: $id) }
	`, map[string]any{"id": state.ID.ValueString()}, nil); err != nil {
		resp.Diagnostics.AddError("Delete notification channel failed", err.Error())
	}
}

func (r *notificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *notificationChannelResource) input(model notificationChannelModel) map[string]any {
	input := map[string]any{
		"name": model.Name.ValueString(),
		"type": model.Type.ValueString(),
	}

	if config := keyValuesInput(model.Config); config != nil {
		input["config"] = config
	}

	return input
}

func (r *notificationChannelResource) readChannel(ctx context.Context, id string) (*notificationChannelModel, error) {
	var out struct {
		NotificationChannel *gqlNotificationChannel `json:"notificationChannel"`
	}

	if err := r.client.Query(ctx, `
		query ($id: ID!) {
			notificationChannel(id: $id) {
				id name type
				config { key value }
			}
		}
	`, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}

	if out.NotificationChannel == nil {
		return nil, nil
	}

	state := notificationChannelFromAPI(*out.NotificationChannel)
	return &state, nil
}

func notificationChannelFromAPI(channel gqlNotificationChannel) notificationChannelModel {
	return notificationChannelModel{
		ID:     types.StringValue(channel.ID),
		Name:   types.StringValue(channel.Name),
		Type:   types.StringValue(channel.Type),
		Config: keyValuesModel(channel.Config),
	}
}
