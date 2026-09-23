package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
				MarkdownDescription: "Channel settings as key/value pairs (`url`, `to`, `routing_key`, ...). Sent as the typed GraphQL input for `type`, such as `pagerduty { routingKey }`.",
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

	input, err := notificationChannelInput(plan)
	if err != nil {
		resp.Diagnostics.AddError("Create notification channel failed", err.Error())
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($input: CreateNotificationChannelInput!) {
			createNotificationChannel(input: $input) {`+notificationChannelSelection+`}
		}
	`, map[string]any{"input": input}, &out); err != nil {
		resp.Diagnostics.AddError("Create notification channel failed", err.Error())
		return
	}

	state := notificationChannelFromAPI(out.CreateNotificationChannel, plan.Config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.readChannel(ctx, state.ID.ValueString(), state.Config)
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

	input, err := notificationChannelInput(plan)
	if err != nil {
		resp.Diagnostics.AddError("Update notification channel failed", err.Error())
		return
	}

	if err := r.client.Query(ctx, `
		mutation ($id: ID!, $input: UpdateNotificationChannelInput!) {
			updateNotificationChannel(id: $id, input: $input) { id }
		}
	`, map[string]any{"id": plan.ID.ValueString(), "input": input}, nil); err != nil {
		resp.Diagnostics.AddError("Update notification channel failed", err.Error())
		return
	}

	refreshed, err := r.readChannel(ctx, plan.ID.ValueString(), plan.Config)
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

const notificationChannelSelection = `
	id
	name
	type
	mail { to host port username password encryption fromAddress fromName }
	slack { webhookUrl }
	microsoftTeams { webhookUrl }
	discord { webhookUrl }
	webhook { url }
	pagerduty { routingKey }
`

type channelConfigField struct {
	graphQL string
	key     string
	aliases []string
	integer bool
}

// Nominal stores channel settings on typed GraphQL objects. Terraform keeps
// the key/value config block and this map translates both directions.
var channelConfigFields = map[string]struct {
	input  string
	fields []channelConfigField
}{
	"mail": {
		input: "mail",
		fields: []channelConfigField{
			{graphQL: "to", key: "to"},
			{graphQL: "host", key: "host"},
			{graphQL: "port", key: "port", integer: true},
			{graphQL: "username", key: "username"},
			{graphQL: "password", key: "password"},
			{graphQL: "encryption", key: "encryption"},
			{graphQL: "fromAddress", key: "from_address"},
			{graphQL: "fromName", key: "from_name"},
		},
	},
	"slack": {
		input: "slack",
		fields: []channelConfigField{
			{graphQL: "webhookUrl", key: "webhook_url", aliases: []string{"url"}},
		},
	},
	"microsoftteams": {
		input: "microsoftTeams",
		fields: []channelConfigField{
			{graphQL: "webhookUrl", key: "webhook_url", aliases: []string{"url"}},
		},
	},
	"discord": {
		input: "discord",
		fields: []channelConfigField{
			{graphQL: "webhookUrl", key: "webhook_url", aliases: []string{"url"}},
		},
	},
	"webhook": {
		input: "webhook",
		fields: []channelConfigField{
			{graphQL: "url", key: "url", aliases: []string{"webhook_url"}},
		},
	},
	"pagerduty": {
		input: "pagerduty",
		fields: []channelConfigField{
			{graphQL: "routingKey", key: "routing_key", aliases: []string{"integration_key"}},
		},
	},
}

func notificationChannelInput(model notificationChannelModel) (map[string]any, error) {
	input := map[string]any{
		"name": model.Name.ValueString(),
		"type": model.Type.ValueString(),
	}

	spec, ok := channelConfigFields[channelTypeKey(model.Type.ValueString())]
	if !ok {
		if len(model.Config) > 0 {
			return nil, fmt.Errorf("unsupported notification channel type %q", model.Type.ValueString())
		}

		return input, nil
	}

	typed := map[string]any{}
	for _, field := range spec.fields {
		value, found := channelConfigValue(model.Config, field)
		if !found {
			continue
		}

		if field.integer {
			port, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("config %s must be an integer", field.key)
			}
			typed[field.graphQL] = port
			continue
		}

		typed[field.graphQL] = value
	}

	if len(typed) > 0 {
		input[spec.input] = typed
	}

	return input, nil
}

func (r *notificationChannelResource) readChannel(ctx context.Context, id string, previous []keyValueModel) (*notificationChannelModel, error) {
	var out struct {
		NotificationChannel *gqlNotificationChannel `json:"notificationChannel"`
	}

	if err := r.client.Query(ctx, `
		query ($id: ID!) {
			notificationChannel(id: $id) {`+notificationChannelSelection+`}
		}
	`, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}

	if out.NotificationChannel == nil {
		return nil, nil
	}

	state := notificationChannelFromAPI(*out.NotificationChannel, previous)
	return &state, nil
}

func notificationChannelFromAPI(channel gqlNotificationChannel, previous []keyValueModel) notificationChannelModel {
	return notificationChannelModel{
		ID:     types.StringValue(channel.ID),
		Name:   types.StringValue(channel.Name),
		Type:   types.StringValue(channel.Type),
		Config: configFromTypedChannel(channel, previous),
	}
}

func channelTypeKey(channelType string) string {
	return strings.ToLower(strings.ReplaceAll(channelType, "_", ""))
}

func channelConfigValue(config []keyValueModel, field channelConfigField) (string, bool) {
	keys := append([]string{field.key}, field.aliases...)
	for _, item := range config {
		if item.Key.IsNull() || item.Key.IsUnknown() || item.Value.IsNull() || item.Value.IsUnknown() {
			continue
		}

		for _, key := range keys {
			if item.Key.ValueString() == key && item.Value.ValueString() != "" {
				return item.Value.ValueString(), true
			}
		}
	}

	return "", false
}

func configFromTypedChannel(channel gqlNotificationChannel, previous []keyValueModel) []keyValueModel {
	spec, ok := channelConfigFields[channelTypeKey(channel.Type)]
	if !ok {
		return nil
	}

	values := typedChannelValues(channel)
	if len(values) == 0 {
		return nil
	}

	used := map[string]bool{}
	var config []keyValueModel
	for _, item := range previous {
		if item.Key.IsNull() || item.Key.IsUnknown() {
			continue
		}

		canonical := canonicalChannelKey(spec.fields, item.Key.ValueString())
		value, found := values[canonical]
		if canonical == "" || !found || used[canonical] {
			continue
		}

		config = append(config, keyValueModel{
			Key:   item.Key,
			Value: types.StringValue(value),
		})
		used[canonical] = true
	}

	for _, field := range spec.fields {
		if used[field.key] {
			continue
		}

		value, found := values[field.key]
		if !found {
			continue
		}

		config = append(config, keyValueModel{
			Key:   types.StringValue(field.key),
			Value: types.StringValue(value),
		})
	}

	return config
}

func canonicalChannelKey(fields []channelConfigField, key string) string {
	for _, field := range fields {
		if key == field.key {
			return field.key
		}

		for _, alias := range field.aliases {
			if key == alias {
				return field.key
			}
		}
	}

	return ""
}

func typedChannelValues(channel gqlNotificationChannel) map[string]string {
	values := map[string]string{}
	add := func(key string, value *string) {
		if value != nil && *value != "" {
			values[key] = *value
		}
	}

	switch channelTypeKey(channel.Type) {
	case "mail":
		if channel.Mail == nil {
			return values
		}
		add("to", channel.Mail.To)
		add("host", channel.Mail.Host)
		add("username", channel.Mail.Username)
		add("password", channel.Mail.Password)
		add("encryption", channel.Mail.Encryption)
		add("from_address", channel.Mail.FromAddress)
		add("from_name", channel.Mail.FromName)
		if channel.Mail.Port != nil {
			values["port"] = strconv.Itoa(*channel.Mail.Port)
		}
	case "slack":
		if channel.Slack != nil {
			add("webhook_url", channel.Slack.WebhookURL)
		}
	case "microsoftteams":
		if channel.MicrosoftTeams != nil {
			add("webhook_url", channel.MicrosoftTeams.WebhookURL)
		}
	case "discord":
		if channel.Discord != nil {
			add("webhook_url", channel.Discord.WebhookURL)
		}
	case "webhook":
		if channel.Webhook != nil {
			add("url", channel.Webhook.URL)
		}
	case "pagerduty":
		if channel.Pagerduty != nil {
			add("routing_key", channel.Pagerduty.RoutingKey)
		}
	}

	return values
}
