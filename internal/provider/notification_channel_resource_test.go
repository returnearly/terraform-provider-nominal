package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNotificationChannelInputUsesTypedPagerdutyField(t *testing.T) {
	t.Parallel()

	input, err := notificationChannelInput(notificationChannelModel{
		Name: types.StringValue("PagerDuty - Identity"),
		Type: types.StringValue("Pagerduty"),
		Config: []keyValueModel{{
			Key:   types.StringValue("routing_key"),
			Value: types.StringValue("R0123456789ABCDEF"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := input["config"]; ok {
		t.Fatal("config must not be sent to GraphQL")
	}

	pagerduty, _ := input["pagerduty"].(map[string]any)
	if pagerduty["routingKey"] != "R0123456789ABCDEF" {
		t.Fatalf("pagerduty: %#v", input["pagerduty"])
	}
}

func TestNotificationChannelInputMapsAliasesAndMailPort(t *testing.T) {
	t.Parallel()

	pagerduty, err := notificationChannelInput(notificationChannelModel{
		Name: types.StringValue("PagerDuty"),
		Type: types.StringValue("Pagerduty"),
		Config: []keyValueModel{{
			Key:   types.StringValue("integration_key"),
			Value: types.StringValue("R0123456789ABCDEF"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	typed, _ := pagerduty["pagerduty"].(map[string]any)
	if typed["routingKey"] != "R0123456789ABCDEF" {
		t.Fatalf("alias: %#v", pagerduty["pagerduty"])
	}

	slack, err := notificationChannelInput(notificationChannelModel{
		Name: types.StringValue("Slack"),
		Type: types.StringValue("Slack"),
		Config: []keyValueModel{{
			Key:   types.StringValue("url"),
			Value: types.StringValue("https://hooks.slack.com/services/T/B/xxx"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	slackTyped, _ := slack["slack"].(map[string]any)
	if slackTyped["webhookUrl"] != "https://hooks.slack.com/services/T/B/xxx" {
		t.Fatalf("slack: %#v", slack["slack"])
	}

	mail, err := notificationChannelInput(notificationChannelModel{
		Name: types.StringValue("Mail"),
		Type: types.StringValue("Mail"),
		Config: []keyValueModel{
			{Key: types.StringValue("to"), Value: types.StringValue("ops@example.com")},
			{Key: types.StringValue("port"), Value: types.StringValue("587")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mailTyped, _ := mail["mail"].(map[string]any)
	if mailTyped["to"] != "ops@example.com" || mailTyped["port"] != 587 {
		t.Fatalf("mail: %#v", mail["mail"])
	}
}

func TestNotificationChannelStateKeepsConfiguredAlias(t *testing.T) {
	t.Parallel()

	routingKey := "R0123456789ABCDEF"
	state := notificationChannelFromAPI(gqlNotificationChannel{
		ID:   "channel-1",
		Name: "PagerDuty",
		Type: "Pagerduty",
		Pagerduty: &gqlPagerdutyConfig{
			RoutingKey: &routingKey,
		},
	}, []keyValueModel{{
		Key:   types.StringValue("integration_key"),
		Value: types.StringValue("old"),
	}})

	if len(state.Config) != 1 {
		t.Fatalf("config: %#v", state.Config)
	}

	if state.Config[0].Key.ValueString() != "integration_key" || state.Config[0].Value.ValueString() != routingKey {
		t.Fatalf("config: %s=%s", state.Config[0].Key.ValueString(), state.Config[0].Value.ValueString())
	}
}

func TestNotificationChannelStateUsesCanonicalKeyWithoutPreviousConfig(t *testing.T) {
	t.Parallel()

	routingKey := "R0123456789ABCDEF"
	state := notificationChannelFromAPI(gqlNotificationChannel{
		ID:   "channel-1",
		Name: "PagerDuty",
		Type: "Pagerduty",
		Pagerduty: &gqlPagerdutyConfig{
			RoutingKey: &routingKey,
		},
	}, nil)

	if len(state.Config) != 1 || state.Config[0].Key.ValueString() != "routing_key" {
		t.Fatalf("config: %#v", state.Config)
	}
}
