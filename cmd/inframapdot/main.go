package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type networkNode struct {
	ID        string
	Label     string
	FillColor string
	Shape     string
}

type networkEdge struct {
	From  string
	To    string
	Label string
	Style string
}

func main() {
	regionLabel := flag.String("region-label", "AWS Region\\n${var.aws_region}", "DOT label for the AWS region cluster.")
	flag.Parse()

	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dot: %v\n", err)
		os.Exit(1)
	}

	output := buildNetworkView(string(inputBytes), *regionLabel)
	fmt.Print(output)
}

func buildNetworkView(rawDOT string, regionLabel string) string {
	nodes := []networkNode{
		{ID: "client.request", Label: "Client\\nrequest", FillColor: "#fff8c5", Shape: "oval"},
		{ID: "aws_internet_gateway.main", Label: "aws_internet_gateway.main", FillColor: "#ddf4ff", Shape: "box"},
		{ID: "aws_route_table.public", Label: "aws_route_table.public", FillColor: "#ddf4ff", Shape: "box"},
		{ID: "aws_route.public_internet", Label: "aws_route.public_internet\\n0.0.0.0/0", FillColor: "#ddf4ff", Shape: "box"},
		{ID: "aws_subnet.public", Label: "aws_subnet.public", FillColor: "#f6f8fa", Shape: "box"},
		{ID: "aws_lb.app", Label: "aws_lb.app\\npublic ALB", FillColor: "#dafbe1", Shape: "box"},
		{ID: "aws_lb_listener.http", Label: "aws_lb_listener.http\\nHTTP :80", FillColor: "#dafbe1", Shape: "box"},
		{ID: "aws_lb_target_group.api", Label: "aws_lb_target_group.api", FillColor: "#dafbe1", Shape: "box"},
		{ID: "aws_subnet.private", Label: "aws_subnet.private", FillColor: "#f6f8fa", Shape: "box"},
		{ID: "aws_ecs_cluster.api", Label: "aws_ecs_cluster.api", FillColor: "#ffeef0", Shape: "box"},
		{ID: "aws_ecs_service.api", Label: "aws_ecs_service.api\\nFargate service", FillColor: "#ffeef0", Shape: "box"},
		{ID: "aws_db_instance.app", Label: "aws_db_instance.app\\nPostgreSQL", FillColor: "#fbefff", Shape: "box"},
	}

	edges := []networkEdge{
		{From: "client.request", To: "aws_internet_gateway.main", Label: "HTTP request"},
		{From: "aws_internet_gateway.main", To: "aws_route.public_internet", Label: "internet ingress"},
		{From: "aws_route.public_internet", To: "aws_route_table.public", Label: "route table entry"},
		{From: "aws_route_table.public", To: "aws_subnet.public", Label: "associated"},
		{From: "aws_subnet.public", To: "aws_lb.app", Label: "public entrypoint"},
		{From: "aws_lb.app", To: "aws_lb_listener.http", Label: "listen"},
		{From: "aws_lb_listener.http", To: "aws_lb_target_group.api", Label: "forward"},
		{From: "aws_lb_target_group.api", To: "aws_subnet.private", Label: "targets"},
		{From: "aws_subnet.private", To: "aws_ecs_service.api", Label: "runs in"},
		{From: "aws_ecs_service.api", To: "aws_ecs_cluster.api", Label: "belongs to", Style: "dashed"},
		{From: "aws_ecs_service.api", To: "aws_db_instance.app", Label: "DB access"},
	}

	var builder strings.Builder
	builder.WriteString("strict digraph G {\n")
	builder.WriteString("\tgraph [ rankdir=LR, compound=true, nodesep=0.55, ranksep=0.85, fontname=\"Arial\" ];\n")
	builder.WriteString("\tnode [ style=\"rounded,filled\", fontname=\"Arial\", fontsize=11, margin=0.12 ];\n")
	builder.WriteString("\tedge [ fontname=\"Arial\", fontsize=10, color=\"#57606a\", arrowsize=0.8 ];\n")
	writeNode(&builder, nodes[0])
	builder.WriteString("\tsubgraph cluster_region {\n")
	builder.WriteString(fmt.Sprintf("\t\tlabel=%s;\n", quoteDOTID(regionLabel)))
	builder.WriteString("\t\tstyle=\"rounded,filled\";\n")
	builder.WriteString("\t\tcolor=\"#b6e3ff\";\n")
	builder.WriteString("\t\tfillcolor=\"#f6fbff\";\n")
	builder.WriteString("\t\tsubgraph cluster_vpc {\n")
	builder.WriteString("\t\t\tlabel=\"VPC\\naws_vpc.main\";\n")
	builder.WriteString("\t\t\tstyle=\"rounded,filled\";\n")
	builder.WriteString("\t\t\tcolor=\"#8cbeff\";\n")
	builder.WriteString("\t\t\tfillcolor=\"#ffffff\";\n")
	writeNode(&builder, findNode(nodes, "aws_internet_gateway.main"))
	builder.WriteString("\t\t\tsubgraph cluster_public_subnet {\n")
	builder.WriteString("\t\t\t\tlabel=\"Public subnet\";\n")
	builder.WriteString("\t\t\t\tstyle=\"rounded,filled\";\n")
	builder.WriteString("\t\t\t\tcolor=\"#a5d6ff\";\n")
	builder.WriteString("\t\t\t\tfillcolor=\"#f6f8fa\";\n")
	for _, id := range []string{
		"aws_subnet.public",
		"aws_route_table.public",
		"aws_route.public_internet",
		"aws_lb.app",
		"aws_lb_listener.http",
		"aws_lb_target_group.api",
	} {
		writeNode(&builder, findNode(nodes, id))
	}
	builder.WriteString("\t\t\t}\n")
	builder.WriteString("\t\t\tsubgraph cluster_private_subnet {\n")
	builder.WriteString("\t\t\t\tlabel=\"Private subnet\";\n")
	builder.WriteString("\t\t\t\tstyle=\"rounded,filled\";\n")
	builder.WriteString("\t\t\t\tcolor=\"#ffb3c7\";\n")
	builder.WriteString("\t\t\t\tfillcolor=\"#fff8f8\";\n")
	for _, id := range []string{
		"aws_subnet.private",
		"aws_ecs_cluster.api",
		"aws_ecs_service.api",
		"aws_db_instance.app",
	} {
		writeNode(&builder, findNode(nodes, id))
	}
	builder.WriteString("\t\t\t}\n")
	builder.WriteString("\t\t}\n")
	builder.WriteString("\t}\n")
	for _, edge := range edges {
		if !hasNode(rawDOT, edge.From) && !strings.HasPrefix(edge.From, "client.") {
			continue
		}
		if !hasNode(rawDOT, edge.To) && !strings.HasPrefix(edge.To, "client.") {
			continue
		}
		writeEdge(&builder, edge)
	}
	builder.WriteString("}\n")
	return builder.String()
}

func writeNode(builder *strings.Builder, node networkNode) {
	builder.WriteString(fmt.Sprintf(
		"\t\t\t\t%s [ label=%s, fillcolor=%s, shape=%s ];\n",
		quoteDOTID(node.ID),
		quoteDOTID(node.Label),
		quoteDOTID(node.FillColor),
		node.Shape,
	))
}

func writeEdge(builder *strings.Builder, edge networkEdge) {
	attributes := []string{}
	if edge.Label != "" {
		attributes = append(attributes, "label="+quoteDOTID(edge.Label))
	}
	if edge.Style != "" {
		attributes = append(attributes, "style="+quoteDOTID(edge.Style))
	}

	attributeText := ""
	if len(attributes) > 0 {
		attributeText = " [ " + strings.Join(attributes, ", ") + " ]"
	}
	builder.WriteString(fmt.Sprintf("\t%s->%s%s;\n", quoteDOTID(edge.From), quoteDOTID(edge.To), attributeText))
}

func findNode(nodes []networkNode, id string) networkNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return networkNode{ID: id, Label: id, FillColor: "#ffffff", Shape: "box"}
}

func hasNode(dot string, id string) bool {
	return strings.Contains(dot, quoteDOTID(id))
}

func quoteDOTID(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}
