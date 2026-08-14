package main

import (
	"strings"
	"testing"
)

func TestBuildNetworkViewShowsRequestPathAndContainment(t *testing.T) {
	rawDOT := `strict digraph G {
	"aws_iam_role.ecs_task" [ shape=ellipse ];
	"aws_secretsmanager_secret.db" [ shape=ellipse ];
	"aws_vpc.main" [ shape=ellipse ];
	"aws_internet_gateway.main" [ shape=ellipse ];
	"aws_route.public_internet" [ shape=ellipse ];
	"aws_route_table.public" [ shape=ellipse ];
	"aws_subnet.public" [ shape=ellipse ];
	"aws_subnet.private" [ shape=ellipse ];
	"aws_lb.app" [ shape=ellipse ];
	"aws_lb_listener.http" [ shape=ellipse ];
	"aws_lb_target_group.api" [ shape=ellipse ];
	"aws_ecs_cluster.api" [ shape=ellipse ];
	"aws_ecs_service.api" [ shape=ellipse ];
	"aws_db_instance.app" [ shape=ellipse ];
}`

	output := buildNetworkView(rawDOT, "AWS Region\\n${var.aws_region}")

	for _, expected := range []string{
		"subgraph cluster_region",
		"subgraph cluster_vpc",
		`label="VPC\naws_vpc.main"`,
		"subgraph cluster_public_subnet",
		"subgraph cluster_private_subnet",
		`"client.request"->"aws_internet_gateway.main"`,
		`"aws_internet_gateway.main"->"aws_route.public_internet"`,
		`"aws_route.public_internet"->"aws_route_table.public"`,
		`"aws_route_table.public"->"aws_subnet.public"`,
		`"aws_subnet.private"->"aws_ecs_service.api"`,
		`"aws_ecs_service.api"->"aws_db_instance.app"`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("buildNetworkView() output missing %q:\n%s", expected, output)
		}
	}
}

func TestBuildNetworkViewHidesIAMAndSecretsManager(t *testing.T) {
	rawDOT := `strict digraph G {
	"aws_iam_role.ecs_task" [ shape=ellipse ];
	"aws_secretsmanager_secret.db" [ shape=ellipse ];
	"aws_vpc.main" [ shape=ellipse ];
}`

	output := buildNetworkView(rawDOT, "AWS Region\\n${var.aws_region}")

	for _, hidden := range []string{"aws_iam_role.ecs_task", "aws_secretsmanager_secret.db", "provider.aws.region"} {
		if strings.Contains(output, hidden) {
			t.Fatalf("buildNetworkView() output contains hidden resource %q:\n%s", hidden, output)
		}
	}
}
