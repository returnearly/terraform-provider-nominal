package provider

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMonitorInputMapsNewFieldsAndDropsGroup(t *testing.T) {
	t.Parallel()

	resource := &monitorResource{}
	input := resource.input(context.Background(), monitorModel{
		Name:            types.StringValue("API health"),
		Description:     types.StringValue("Owned by payments."),
		Tags:            stringListValue([]string{"prod", "critical"}),
		Group:           types.StringValue("legacy-group"),
		Type:            types.StringValue("Http"),
		Enabled:         types.BoolValue(true),
		IntervalSeconds: types.Int64Value(60),
		TimeoutSeconds:  types.Int64Value(10),
		IPFamily:        types.StringValue("Any"),
		Target:          types.StringValue("https://example.com/health"),
		Method:          types.StringValue("GET"),
		RequestBody:     types.StringValue(""),
		DNSQueryName:    types.StringNull(),
		DNSQueryType:    types.StringNull(),
		FollowRedirects: types.BoolValue(true),
		VerifyTLS:       types.BoolValue(true),
		ProxyURL:        types.StringValue("socks5h://127.0.0.1:1080"),
		RetentionDays:   types.Int64Value(30),
		Conditions:      stringListValue([]string{"[STATUS] == 200", "[DOMAIN_EXPIRATION] > 720h"}),
		ProbeIDs:        stringListValue([]string{"probe-1"}),
		RequestHeaders: []keyValueModel{
			{Key: types.StringValue("X-Token"), Value: types.StringValue("abc")},
		},
	})

	if _, ok := input["group"]; ok {
		t.Fatal("group must not be sent to GraphQL")
	}

	if input["description"] != "Owned by payments." {
		t.Fatalf("description: %#v", input["description"])
	}

	if input["proxyUrl"] != "socks5h://127.0.0.1:1080" {
		t.Fatalf("proxyUrl: %#v", input["proxyUrl"])
	}

	tags, _ := input["tags"].([]string)
	if !reflect.DeepEqual(tags, []string{"prod", "critical"}) {
		t.Fatalf("tags: %#v", input["tags"])
	}

	conditions, _ := input["conditions"].([]string)
	if !reflect.DeepEqual(conditions, []string{"[STATUS] == 200", "[DOMAIN_EXPIRATION] > 720h"}) {
		t.Fatalf("conditions: %#v", input["conditions"])
	}

	headers, _ := input["requestHeaders"].([]map[string]string)
	if len(headers) != 1 || headers[0]["key"] != "X-Token" {
		t.Fatalf("requestHeaders: %#v", input["requestHeaders"])
	}
}

func TestMonitorInputSendsGroupAsTagWhenTagsOmitted(t *testing.T) {
	t.Parallel()

	resource := &monitorResource{}
	input := resource.input(context.Background(), monitorModel{
		Name:            types.StringValue("API"),
		Tags:            types.ListNull(types.StringType),
		Group:           types.StringValue("prod"),
		Type:            types.StringValue("Dns"),
		Enabled:         types.BoolValue(true),
		IntervalSeconds: types.Int64Value(300),
		TimeoutSeconds:  types.Int64Value(10),
		IPFamily:        types.StringValue("Ipv4"),
		Target:          types.StringValue("1.1.1.1"),
		DNSQueryName:    types.StringValue("example.com"),
		DNSQueryType:    types.StringValue("A"),
		FollowRedirects: types.BoolValue(true),
		VerifyTLS:       types.BoolValue(true),
		RetentionDays:   types.Int64Value(30),
	})

	tags, _ := input["tags"].([]string)
	if !reflect.DeepEqual(tags, []string{"prod"}) {
		t.Fatalf("expected group to map to tags, got %#v", input["tags"])
	}

	if input["dnsQueryName"] != "example.com" || input["dnsQueryType"] != "A" {
		t.Fatalf("dns fields: %#v %#v", input["dnsQueryName"], input["dnsQueryType"])
	}
}

func TestStatusPageInputIncludesMonitors(t *testing.T) {
	t.Parallel()

	resource := &statusPageResource{}
	input := resource.input(statusPageModel{
		Name:           types.StringValue("Acme Status"),
		Slug:           types.StringValue("acme"),
		Theme:          types.StringValue("Dark"),
		Published:      types.BoolValue(true),
		ShowTargets:    types.BoolValue(false),
		RefreshSeconds: types.Int64Value(30),
		Password:       types.StringValue("secret"),
		CustomDomain:   types.StringValue("status.acme.test"),
		Monitors: []statusPageMonitorModel{
			{
				MonitorID:  types.StringValue("mon-1"),
				PublicName: types.StringValue("API"),
			},
		},
	})

	if input["slug"] != "acme" || input["theme"] != "Dark" || input["password"] != "secret" {
		t.Fatalf("status page input: %#v", input)
	}

	monitors, _ := input["monitors"].([]map[string]any)
	if len(monitors) != 1 || monitors[0]["monitorId"] != "mon-1" || monitors[0]["publicName"] != "API" {
		t.Fatalf("monitors: %#v", input["monitors"])
	}
}

func TestMaintenanceWindowInput(t *testing.T) {
	t.Parallel()

	resource := &maintenanceWindowResource{}
	input := resource.input(context.Background(), maintenanceWindowModel{
		Title:        types.StringValue("Database upgrade"),
		Message:      types.StringValue("Upgrading Postgres."),
		StartsAt:     types.StringValue("2026-08-21 02:00:00"),
		EndsAt:       types.StringValue("2026-08-21 04:00:00"),
		AppliesToAll: types.BoolValue(false),
		MonitorIDs:   stringListValue([]string{"mon-1"}),
	})

	if input["title"] != "Database upgrade" || input["startsAt"] != "2026-08-21 02:00:00" {
		t.Fatalf("window input: %#v", input)
	}

	ids, _ := input["monitorIds"].([]string)
	if !reflect.DeepEqual(ids, []string{"mon-1"}) {
		t.Fatalf("monitorIds: %#v", input["monitorIds"])
	}
}

func TestMonitorSelectionCoversNewAPIFields(t *testing.T) {
	t.Parallel()

	for _, field := range []string{
		"description",
		"tags",
		"dns_query_name",
		"dns_query_type",
		"proxy_url",
		"heartbeatUrl",
		"heartbeatStartUrl",
		"heartbeatFinishUrl",
		"heartbeatErrorUrl",
		"statusBadgeUrl",
		"uptime {",
	} {
		if !strings.Contains(monitorSelection, field) {
			t.Fatalf("monitor selection missing %q", field)
		}
	}
}
