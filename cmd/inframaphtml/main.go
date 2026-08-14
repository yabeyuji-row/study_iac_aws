package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type pageData struct {
	Title       string
	SVG         template.HTML
	Description template.JS
	Variables   template.JS
	Schema      template.JS
	Context     template.JS
}

type terraformContext struct {
	Locals    map[string]any            `json:"locals"`
	Resources map[string]map[string]any `json:"resources"`
}

type providerSchemasFile struct {
	ProviderSchemas map[string]providerSchema `json:"provider_schemas"`
}

type providerSchema struct {
	ResourceSchemas map[string]resourceSchema `json:"resource_schemas"`
}

type resourceSchema struct {
	Block blockSchema `json:"block"`
}

type blockSchema struct {
	Attributes map[string]attributeSchema `json:"attributes"`
}

type attributeSchema struct {
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

func main() {
	svgPath := flag.String("svg", "", "Path to the Graphviz-generated SVG file.")
	descriptionPath := flag.String("description", "", "Path to the InfraMap description JSON file.")
	varsPath := flag.String("vars", "", "Path to the Terraform tfvars file used to resolve var references.")
	terraformDir := flag.String("terraform-dir", "", "Path to the Terraform configuration directory.")
	schemaPath := flag.String("schema", "", "Path to Terraform provider schema JSON.")
	outputPath := flag.String("out", "", "Path to write the interactive HTML file.")
	title := flag.String("title", "Terraform InfraMap", "HTML page title.")
	flag.Parse()

	if *svgPath == "" || *descriptionPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "-svg, -description, and -out are required")
		os.Exit(2)
	}

	svgBytes, err := os.ReadFile(*svgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read svg: %v\n", err)
		os.Exit(1)
	}

	svg, err := copyLocalImages(string(svgBytes), filepath.Dir(*outputPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "copy svg images: %v\n", err)
		os.Exit(1)
	}

	descriptionBytes, err := os.ReadFile(*descriptionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read description: %v\n", err)
		os.Exit(1)
	}

	var description map[string]any
	if err := json.Unmarshal(descriptionBytes, &description); err != nil {
		fmt.Fprintf(os.Stderr, "parse description json: %v\n", err)
		os.Exit(1)
	}
	pruneHiddenDescription(description)
	if err := writeIndentedJSON(*descriptionPath, description); err != nil {
		fmt.Fprintf(os.Stderr, "format description json: %v\n", err)
		os.Exit(1)
	}

	var variables map[string]any
	if *varsPath != "" {
		varsBytes, err := os.ReadFile(*varsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read vars: %v\n", err)
			os.Exit(1)
		}

		variables = parseTfvars(string(varsBytes))
	} else {
		variables = map[string]any{}
	}

	context := terraformContext{
		Locals:    map[string]any{},
		Resources: map[string]map[string]any{},
	}
	if *terraformDir != "" {
		context, err = loadTerraformContext(*terraformDir, variables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse terraform context: %v\n", err)
			os.Exit(1)
		}
	}
	pruneHiddenResourceContext(context)

	schema := map[string]map[string]attributeSchema{}
	if *schemaPath != "" {
		schemaBytes, err := os.ReadFile(*schemaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
			os.Exit(1)
		}

		schema, err = loadAttributeSchema(schemaBytes, description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse schema: %v\n", err)
			os.Exit(1)
		}
	}

	descriptionJSON, err := json.Marshal(description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode description json: %v\n", err)
		os.Exit(1)
	}

	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode variables json: %v\n", err)
		os.Exit(1)
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode schema json: %v\n", err)
		os.Exit(1)
	}

	contextJSON, err := json.Marshal(context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode context json: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output directory: %v\n", err)
		os.Exit(1)
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	data := pageData{
		Title:       *title,
		SVG:         template.HTML(svg),
		Description: template.JS(descriptionJSON),
		Variables:   template.JS(variablesJSON),
		Schema:      template.JS(schemaJSON),
		Context:     template.JS(contextJSON),
	}

	if err := pageTemplate.Execute(outputFile, data); err != nil {
		fmt.Fprintf(os.Stderr, "write html: %v\n", err)
		os.Exit(1)
	}
}

func writeIndentedJSON(path string, value any) error {
	indented, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	indented = append(indented, '\n')
	return os.WriteFile(path, indented, 0o644)
}

func copyLocalImages(svg string, outputDir string) (string, error) {
	hrefPattern := regexp.MustCompile(`\b((?:xlink:)?href)="([^"]+)"`)
	var firstErr error

	result := hrefPattern.ReplaceAllStringFunc(svg, func(match string) string {
		if firstErr != nil {
			return match
		}

		parts := hrefPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		href := parts[2]
		if strings.HasPrefix(href, "data:") || strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			return match
		}

		if !filepath.IsAbs(href) {
			return match
		}

		relativeAssetPath := localAssetPath(href)
		destinationPath := filepath.Join(outputDir, filepath.FromSlash(relativeAssetPath))
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
			firstErr = fmt.Errorf("create asset directory for %q: %w", destinationPath, err)
			return match
		}

		imageBytes, err := os.ReadFile(href)
		if err != nil {
			firstErr = fmt.Errorf("read referenced image %q: %w", href, err)
			return match
		}

		if err := os.WriteFile(destinationPath, imageBytes, 0o644); err != nil {
			firstErr = fmt.Errorf("write asset %q: %w", destinationPath, err)
			return match
		}

		return fmt.Sprintf(`%s="%s"`, parts[1], relativeAssetPath)
	})

	if firstErr != nil {
		return "", firstErr
	}

	return result, nil
}

func localAssetPath(sourcePath string) string {
	const marker = "/assets/"

	if index := strings.Index(filepath.ToSlash(sourcePath), marker); index >= 0 {
		return "assets/" + filepath.ToSlash(sourcePath)[index+len(marker):]
	}

	return "assets/" + filepath.Base(sourcePath)
}

func loadTerraformContext(terraformDir string, variables map[string]any) (terraformContext, error) {
	context := terraformContext{
		Locals:    map[string]any{},
		Resources: map[string]map[string]any{},
	}

	entries, err := os.ReadDir(terraformDir)
	if err != nil {
		return context, err
	}

	var filenames []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".tf" {
			continue
		}
		filenames = append(filenames, filepath.Join(terraformDir, entry.Name()))
	}
	sort.Strings(filenames)

	localPattern := regexp.MustCompile(`^\s*locals\s*\{`)
	resourcePattern := regexp.MustCompile(`^\s*resource\s+"([^"]+)"\s+"([^"]+)"\s*\{`)

	for _, filename := range filenames {
		contentBytes, err := os.ReadFile(filename)
		if err != nil {
			return context, err
		}

		lines := strings.Split(string(contentBytes), "\n")
		for index := 0; index < len(lines); index++ {
			line := lines[index]
			if localPattern.MatchString(line) {
				body, nextIndex := collectHCLBlock(lines, index)
				for name, value := range parseHCLAssignments(body) {
					context.Locals[name] = value
				}
				index = nextIndex
				continue
			}

			matches := resourcePattern.FindStringSubmatch(line)
			if len(matches) == 3 {
				body, nextIndex := collectHCLBlock(lines, index)
				address := matches[1] + "." + matches[2]
				context.Resources[address] = parseHCLAssignments(body)
				index = nextIndex
			}
		}
	}

	if value, ok := context.Locals["common_tags"].(string); !ok || strings.HasPrefix(value, "merge(") {
		if tags, ok := variables["tags"]; ok {
			context.Locals["common_tags"] = tags
		}
	}

	return context, nil
}

func pruneHiddenResourceContext(context terraformContext) {
	for resourceName := range context.Resources {
		if !isVisibleNetworkResource(resourceName) || isHiddenResource(resourceName) {
			delete(context.Resources, resourceName)
			continue
		}

		pruneHiddenAttributes(context.Resources[resourceName])
	}
}

func pruneHiddenDescription(description map[string]any) {
	for resourceName, rawAttributes := range description {
		if !isVisibleNetworkResource(resourceName) || isHiddenResource(resourceName) {
			delete(description, resourceName)
			continue
		}

		attributes, ok := rawAttributes.(map[string]any)
		if !ok {
			continue
		}
		pruneHiddenAttributes(attributes)
	}
}

func pruneHiddenAttributes(attributes map[string]any) {
	for key, value := range attributes {
		if isHiddenAttribute(key) || containsHiddenReference(value) {
			delete(attributes, key)
		}
	}
}

func isHiddenResource(resourceName string) bool {
	return strings.HasPrefix(resourceName, "aws_iam_") ||
		strings.HasPrefix(resourceName, "aws_secretsmanager_") ||
		strings.HasPrefix(resourceName, "random_password.")
}

func isVisibleNetworkResource(resourceName string) bool {
	switch resourceName {
	case "aws_db_instance.app",
		"aws_ecs_cluster.api",
		"aws_ecs_service.api",
		"aws_internet_gateway.main",
		"aws_lb.app",
		"aws_lb_listener.http",
		"aws_lb_target_group.api",
		"aws_route.public_internet",
		"aws_route_table.public",
		"aws_route_table_association.public",
		"aws_subnet.private",
		"aws_subnet.public",
		"aws_vpc.main":
		return true
	default:
		return false
	}
}

func isHiddenAttribute(attributeName string) bool {
	switch attributeName {
	case "password", "secret_string", "assume_role_policy", "policy", "policy_arn", "role", "task_role_arn", "execution_role_arn":
		return true
	default:
		return false
	}
}

func containsHiddenReference(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.Contains(typed, "aws_iam_") ||
			strings.Contains(typed, "aws_secretsmanager_") ||
			strings.Contains(typed, "random_password.")
	case []any:
		for _, item := range typed {
			if containsHiddenReference(item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range typed {
			if containsHiddenReference(item) {
				return true
			}
		}
	}
	return false
}

func collectHCLBlock(lines []string, startIndex int) ([]string, int) {
	depth := countDelimiters(lines[startIndex], '{', '}')
	var body []string

	for index := startIndex + 1; index < len(lines); index++ {
		line := lines[index]
		depth += countDelimiters(line, '{', '}')
		if depth <= 0 {
			return body, index
		}
		body = append(body, line)
	}

	return body, len(lines) - 1
}

func parseHCLAssignments(lines []string) map[string]any {
	values := make(map[string]any)

	for index := 0; index < len(lines); index++ {
		line := stripComment(strings.TrimSpace(lines[index]))
		if line == "" || !strings.Contains(line, "=") || !isTopLevelAssignment(lines, index) {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		name := strings.TrimSpace(parts[0])
		rawValue := strings.TrimSpace(parts[1])
		for !isCompleteHCLValue(rawValue) && index+1 < len(lines) {
			index++
			rawValue += "\n" + stripComment(strings.TrimSpace(lines[index]))
		}

		if name != "" {
			values[name] = parseHCLValue(rawValue)
		}
	}

	return values
}

func isTopLevelAssignment(lines []string, lineIndex int) bool {
	depth := 0
	for index := 0; index < lineIndex; index++ {
		depth += countDelimiters(lines[index], '{', '}')
		depth += countDelimiters(lines[index], '[', ']')
		depth += countDelimiters(lines[index], '(', ')')
	}

	return depth == 0
}

func isCompleteHCLValue(value string) bool {
	return strings.Count(value, "[") == strings.Count(value, "]") &&
		strings.Count(value, "{") == strings.Count(value, "}") &&
		strings.Count(value, "(") == strings.Count(value, ")")
}

func parseHCLValue(value string) any {
	value = strings.TrimSpace(strings.TrimSuffix(value, ","))
	if strings.HasPrefix(value, "var.") || strings.HasPrefix(value, "local.") || strings.HasPrefix(value, "aws_") || strings.HasPrefix(value, "random_") {
		return "${" + value + "}"
	}

	return parseTfvarsValue(value)
}

func countDelimiters(line string, open rune, close rune) int {
	depth := 0
	inString := false
	escaped := false

	for _, char := range line {
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' && inString {
			escaped = true
			continue
		}
		if char == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if char == open {
			depth++
		}
		if char == close {
			depth--
		}
	}

	return depth
}

func loadAttributeSchema(schemaBytes []byte, description map[string]any) (map[string]map[string]attributeSchema, error) {
	var schemaFile providerSchemasFile
	if err := json.Unmarshal(schemaBytes, &schemaFile); err != nil {
		return nil, err
	}

	resourceTypes := make(map[string]map[string]struct{})
	for resourceName, rawAttributes := range description {
		attributes, ok := rawAttributes.(map[string]any)
		if !ok {
			continue
		}

		resourceType := terraformResourceType(resourceName)
		if resourceType == "" {
			continue
		}

		if _, ok := resourceTypes[resourceType]; !ok {
			resourceTypes[resourceType] = map[string]struct{}{}
		}
		for attributeName := range attributes {
			resourceTypes[resourceType][attributeName] = struct{}{}
		}
	}

	result := make(map[string]map[string]attributeSchema, len(resourceTypes))
	for _, provider := range schemaFile.ProviderSchemas {
		for resourceType, requestedAttributes := range resourceTypes {
			resource, ok := provider.ResourceSchemas[resourceType]
			if !ok {
				continue
			}
			if _, ok := result[resourceType]; !ok {
				result[resourceType] = map[string]attributeSchema{}
			}
			for attributeName := range requestedAttributes {
				attribute, ok := resource.Block.Attributes[attributeName]
				if !ok {
					continue
				}
				result[resourceType][attributeName] = attributeSchema{
					Required:    attribute.Required,
					Description: firstSentence(attribute.Description),
				}
			}
		}
	}

	return result, nil
}

func terraformResourceType(resourceName string) string {
	lastDot := strings.LastIndex(resourceName, ".")
	if lastDot <= 0 {
		return ""
	}

	return resourceName[:lastDot]
}

func firstSentence(description string) string {
	description = strings.Join(strings.Fields(description), " ")
	if description == "" {
		return ""
	}

	for _, separator := range []string{". ", "\n"} {
		if index := strings.Index(description, separator); index > 0 {
			return strings.TrimSpace(description[:index+1])
		}
	}

	return description
}

func parseTfvars(content string) map[string]any {
	values := make(map[string]any)
	lines := strings.Split(content, "\n")

	for index := 0; index < len(lines); index++ {
		line := stripComment(strings.TrimSpace(lines[index]))
		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		name := strings.TrimSpace(parts[0])
		rawValue := strings.TrimSpace(parts[1])
		for !isCompleteTfvarsValue(rawValue) && index+1 < len(lines) {
			index++
			rawValue += "\n" + stripComment(strings.TrimSpace(lines[index]))
		}

		if name != "" {
			values[name] = parseTfvarsValue(rawValue)
		}
	}

	return values
}

func stripComment(line string) string {
	inString := false
	escaped := false

	for index, char := range line {
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' && inString {
			escaped = true
			continue
		}
		if char == '"' {
			inString = !inString
			continue
		}
		if !inString && char == '#' {
			return strings.TrimSpace(line[:index])
		}
	}

	return line
}

func isCompleteTfvarsValue(value string) bool {
	return strings.Count(value, "[") == strings.Count(value, "]") &&
		strings.Count(value, "{") == strings.Count(value, "}")
}

func parseTfvarsValue(value string) any {
	value = strings.TrimSpace(strings.TrimSuffix(value, ","))
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		var decoded string
		if err := json.Unmarshal([]byte(value), &decoded); err == nil {
			return decoded
		}
		return strings.Trim(value, "\"")
	}

	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return parseTfvarsList(value)
	}
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		return parseTfvarsMap(value)
	}

	var number json.Number
	if err := json.Unmarshal([]byte(value), &number); err == nil {
		if intValue, err := number.Int64(); err == nil {
			return intValue
		}
		if floatValue, err := number.Float64(); err == nil {
			return floatValue
		}
	}

	return value
}

func parseTfvarsList(value string) []any {
	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	if body == "" {
		return []any{}
	}

	var values []any
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, ","))
		if line == "" {
			continue
		}
		values = append(values, parseTfvarsValue(line))
	}

	return values
}

func parseTfvarsMap(value string) map[string]any {
	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "{"), "}"))
	values := make(map[string]any)
	if body == "" {
		return values
	}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, ","))
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.Trim(strings.TrimSpace(parts[0]), "\"")
		values[key] = parseTfvarsValue(strings.TrimSpace(parts[1]))
	}

	return values
}

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="inframap.css">
</head>
<body>
  <header>
    <h1>{{ .Title }}</h1>
    <input class="search" id="search" type="search" placeholder="Filter resources">
  </header>
  <main>
    <div class="map-pane">
      <div class="canvas" id="canvas">{{ .SVG }}</div>
    </div>
    <div class="details-pane" id="details-pane">
      <div class="details-head">
        <h2 class="details-title" id="details-title">Resource details</h2>
        <button class="close" id="close" type="button" aria-label="Close">&times;</button>
      </div>
      <div class="details-body" id="details">
        <div class="placeholder">Select a resource to inspect its configuration.</div>
      </div>
    </div>
  </main>
  <script>
    const descriptions = {{ .Description }};
    const variables = {{ .Variables }};
    const providerSchema = {{ .Schema }};
    const terraformContext = {{ .Context }};
    const nodes = Array.from(document.querySelectorAll('svg g.node'));
    const title = document.getElementById('details-title');
    const details = document.getElementById('details');
    const search = document.getElementById('search');
    const closeButton = document.getElementById('close');
    let activeNode = null;

    function nodeName(node) {
      return node.querySelector('title')?.textContent?.trim() || '';
    }

    function renderCellValue(value) {
      if (value === null) {
        return '<code>null</code>';
      }

      if (typeof value === 'object') {
        return '<pre>' + escapeHTML(JSON.stringify(value, null, 2)) + '</pre>';
      }

      return '<code>' + escapeHTML(String(value)) + '</code>';
    }

    function renderResolvedValue(resourceName, key, original) {
      const resourceAttributes = terraformContext.resources[resourceName] || {};
      const source = Object.prototype.hasOwnProperty.call(resourceAttributes, key)
        ? resourceAttributes[key]
        : original;
      const resolved = resolveValue(source);
      if (JSON.stringify(original) === JSON.stringify(resolved)) {
        return '<span class="no-value">-</span>';
      }

      return typeof resolved === 'object'
        ? '<pre>' + escapeHTML(JSON.stringify(resolved, null, 2)) + '</pre>'
        : '<code>' + escapeHTML(String(resolved)) + '</code>';
    }

    function resourceType(resourceName) {
      const lastDot = resourceName.lastIndexOf('.');
      return lastDot > 0 ? resourceName.slice(0, lastDot) : resourceName;
    }

    function attributeMeta(resourceName, key) {
      return providerSchema[resourceType(resourceName)]?.[key] || {};
    }

    function requiredLabel(meta) {
      return meta.required ? '必須' : '任意';
    }

    const attributeSummaries = {
      allocated_storage: '初期ストレージ容量',
      apply_immediately: '変更を即時反映するか',
      auto_minor_version_upgrade: 'マイナーバージョン自動更新',
      backup_retention_period: '自動バックアップ保持日数',
      backup_window: 'バックアップ実行時間帯',
      cluster: '所属する ECS クラスター',
      copy_tags_to_snapshot: 'スナップショットへタグをコピー',
      cidr_block: '割り当てる CIDR ブロック',
      db_name: '作成するデータベース名',
      db_subnet_group_name: 'RDS を配置するサブネットグループ',
      deletion_protection: '削除保護',
      destination_cidr_block: '宛先 CIDR ブロック',
      deployment_circuit_breaker: 'ECS デプロイ失敗時の制御',
      deployment_maximum_percent: 'デプロイ時の最大稼働率',
      deployment_minimum_healthy_percent: 'デプロイ時の最小正常率',
      depends_on: '明示的な依存関係',
      desired_count: '起動したいタスク数',
      drop_invalid_header_fields: '不正な HTTP ヘッダーの破棄',
      enabled_cloudwatch_logs_exports: 'CloudWatch Logs へ出すログ種別',
      engine: 'DB エンジン',
      engine_version: 'DB エンジンバージョン',
      final_snapshot_identifier: '最終スナップショット名',
      health_check_grace_period_seconds: 'ヘルスチェック猶予秒数',
      iam_database_authentication_enabled: 'IAM DB 認証の有効化',
      identifier: 'RDS インスタンス識別子',
      instance_class: 'RDS インスタンスサイズ',
      internal: '内部向けロードバランサーか',
      launch_type: 'ECS 起動タイプ',
      load_balancer: 'ECS サービスに接続するロードバランサー',
      load_balancer_type: 'ロードバランサー種別',
      maintenance_window: 'メンテナンス時間帯',
      max_allocated_storage: '自動拡張の最大ストレージ容量',
      multi_az: 'Multi-AZ 配置',
      name: 'リソース名',
      network_configuration: 'ECS タスクのネットワーク設定',
      parameter_group_name: 'RDS パラメータグループ',
      port: '待ち受けポート',
      publicly_accessible: 'インターネットから到達可能か',
      region: 'AWS リージョン',
      route_table_id: '関連付けるルートテーブル',
      security_groups: '関連付けるセキュリティグループ',
      skip_final_snapshot: '削除時に最終スナップショットを省略',
      source: '設定元',
      storage_encrypted: 'ストレージ暗号化',
      storage_type: 'ストレージ種別',
      subnets: '配置先サブネット',
      subnet_id: '関連付けるサブネット',
      tags: '付与するタグ',
      task_definition: '利用する ECS タスク定義',
      username: 'DB ユーザー名',
      vpc_security_group_ids: 'RDS に関連付けるセキュリティグループ'
    };

    function attributeSummary(key, meta) {
      return meta.description || attributeSummaries[key] || 'Terraform 属性';
    }

    function renderDetailItem(resourceName, key, value) {
      const meta = attributeMeta(resourceName, key);
      const summary = '<span class="item-summary">' + escapeHTML(attributeSummary(key, meta)) + '</span>';

      return '<section class="detail-item">' +
        '<div class="item-head">' +
          '<span class="required-badge ' + (meta.required ? 'is-required' : 'is-optional') + '">' + requiredLabel(meta) + '</span>' +
          '<div class="item-name"><strong>' + escapeHTML(key) + '</strong>' + summary + '</div>' +
        '</div>' +
        '<div class="item-values">' +
          '<div class="item-value"><span>記述内容</span>' + renderCellValue(value) + '</div>' +
          '<div class="item-value actual-value"><span>実値</span>' + renderResolvedValue(resourceName, key, value) + '</div>' +
        '</div>' +
      '</section>';
    }

    function resolveValue(value) {
      if (Array.isArray(value)) {
        return value.map(resolveValue);
      }

      if (value && typeof value === 'object') {
        return Object.fromEntries(Object.entries(value).map(([key, current]) => [key, resolveValue(current)]));
      }

      if (typeof value !== 'string') {
        return value;
      }

      const exactInterpolation = value.match(/^\$\{([^}]+)\}$/);
      const exactExpression = exactInterpolation ? exactInterpolation[1] : value;
      if (/^(var|local|aws_|random_)/.test(exactExpression) && !/\s/.test(exactExpression)) {
        const resolved = resolveReference(exactExpression);
        if (resolved !== undefined) {
          return resolved;
        }
      }

      return value.replace(/\$\{([^}]+)\}/g, (match, expression) => {
        const resolved = resolveReference(expression);
        if (resolved === undefined) {
          return match;
        }
        return typeof resolved === 'object' ? JSON.stringify(resolved) : String(resolved);
      }).replace(/\b((?:aws|random)_[A-Za-z0-9_]+\.[A-Za-z0-9_]+(?:\.[A-Za-z0-9_]+)?)\b/g, (match, expression) => {
        const resolved = resolveReference(expression);
        if (resolved === undefined) {
          return match;
        }
        return typeof resolved === 'object' ? JSON.stringify(resolved) : String(resolved);
      });
    }

    function resolveReference(expression) {
      const variable = expression.match(/^var\.([A-Za-z0-9_]+)$/);
      if (variable && Object.prototype.hasOwnProperty.call(variables, variable[1])) {
        return variables[variable[1]];
      }

      const local = expression.match(/^local\.([A-Za-z0-9_]+)$/);
      if (local && Object.prototype.hasOwnProperty.call(terraformContext.locals, local[1])) {
        return resolveValue(terraformContext.locals[local[1]]);
      }

      const collection = expression.match(/^([A-Za-z0-9_]+\.[A-Za-z0-9_]+)$/);
      if (collection) {
        return resolveResourceCollection(collection[1]);
      }

      const resource = expression.match(/^([A-Za-z0-9_]+\.[A-Za-z0-9_]+)\.([A-Za-z0-9_]+)$/);
      if (!resource) {
        return undefined;
      }

      const address = resource[1];
      const attribute = resource[2];
      const attributes = terraformContext.resources[address];
      if (!attributes) {
        return undefined;
      }

      if (Object.prototype.hasOwnProperty.call(attributes, attribute)) {
        return resolveValue(attributes[attribute]);
      }

      const representativeKey = ['name', 'identifier', 'family', 'repository', 'description'].find((key) => Object.prototype.hasOwnProperty.call(attributes, key));
      if (!representativeKey) {
        return undefined;
      }

      return resolveValue(attributes[representativeKey]);
    }

    function resolveResourceCollection(address) {
      if (address === 'aws_subnet.private') {
        return subnetNames('private');
      }
      if (address === 'aws_subnet.public') {
        return subnetNames('public');
      }

      const attributes = terraformContext.resources[address];
      if (!attributes) {
        return undefined;
      }

      const representativeKey = ['name', 'identifier', 'family', 'repository', 'description'].find((key) => Object.prototype.hasOwnProperty.call(attributes, key));
      return representativeKey ? resolveValue(attributes[representativeKey]) : undefined;
    }

    function subnetNames(scope) {
      const prefix = resolveReference('local.name_prefix');
      const count = Number(variables.az_count || (variables.availability_zone_names || []).length || 0);
      if (!prefix || !count) {
        return undefined;
      }

      return Array.from({ length: count }, (_, index) => prefix + '-' + scope + '-' + (index + 1));
    }

    function escapeHTML(value) {
      return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#039;');
    }

    function showDetails(node) {
      const name = nodeName(node);
      activeNode = node;
      nodes.forEach((current) => current.classList.toggle('is-selected', current === node));
      title.textContent = name;

      const data = descriptions[name] || terraformContext.resources[name];
      if (!data || Object.keys(data).length === 0) {
        details.innerHTML = '<p class="empty">No description data.</p>';
      } else {
        const items = Object.entries(data)
          .sort(([left], [right]) => left.localeCompare(right))
          .map(([key, value]) => renderDetailItem(name, key, value))
          .join('');
        details.innerHTML = '<div class="detail-list">' + items + '</div>';
      }

    }

    function hideDetails() {
      activeNode = null;
      nodes.forEach((node) => node.classList.remove('is-selected'));
      title.textContent = 'Resource details';
      details.innerHTML = '<div class="placeholder">Select a resource to inspect its configuration.</div>';
    }

    nodes.forEach((node) => {
      const name = nodeName(node);
      node.setAttribute('tabindex', '0');
      node.setAttribute('role', 'button');
      node.setAttribute('aria-label', name);
      node.addEventListener('click', () => showDetails(node));
      node.addEventListener('keydown', (event) => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          showDetails(node);
        }
      });
    });

    search.addEventListener('input', () => {
      const query = search.value.trim().toLowerCase();
      nodes.forEach((node) => {
        const matched = nodeName(node).toLowerCase().includes(query);
        node.style.opacity = query && !matched ? '0.18' : '1';
      });
    });

    document.addEventListener('keydown', (event) => {
      if (event.key === 'Escape') {
        hideDetails();
      }
    });
    closeButton.addEventListener('click', hideDetails);
  </script>
</body>
</html>
`))
