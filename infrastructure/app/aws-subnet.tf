# 手動編集: 可 - ネットワークリソースをここに定義する。
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  #checkov:skip=CKV2_AWS_11: VPC Flow Logs は 7週目の範囲外で、後続の observability フェーズで扱う。
  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-vpc"
    }
  )
}

# VPC 作成時に自動で作られる default security group を閉じる。
# 明示的に作成した ALB/ECS/RDS 用 security group だけを使う。
resource "aws_default_security_group" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-default-sg"
      Role = "unused-default"
    }
  )
}

resource "aws_subnet" "public" {
  # local.public_subnets の要素ごとにパブリックサブネットを1つ作成する。
  # 例: 2 AZ を使う場合は、AZ ごとに1つずつサブネットが作られる。
  for_each = local.public_subnets

  # 7-1-4 で作成した VPC の中にサブネットを配置する。
  vpc_id = aws_vpc.main.id
  # each.value は local.public_subnets の現在の要素を指す。
  # cidr はサブネットに割り当てる IP アドレス範囲。
  cidr_block = each.value.cidr
  # az はサブネットを配置するアベイラビリティゾーン。
  availability_zone = each.value.az
  # ALB は subnet 側の自動 public IP 割り当てを必要としないため、既定では無効にする。
  map_public_ip_on_launch = false

  tags = merge(
    local.common_tags,
    {
      # AWS コンソールで見分けやすいサブネット名。
      Name = each.value.name
      # パブリック/プライベートの用途をタグで明示する。
      Role = "public"
    }
  )
}

resource "aws_subnet" "private" {
  for_each = local.private_subnets

  vpc_id                  = aws_vpc.main.id
  cidr_block              = each.value.cidr
  availability_zone       = each.value.az
  map_public_ip_on_launch = false

  tags = merge(
    local.common_tags,
    {
      Name = each.value.name
      Role = "private"
    }
  )
}

# VPC からインターネットへ出入りするための Internet Gateway。
# public subnet の default route から参照する。
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-igw"
    }
  )
}

# public subnet 用の route table。
# private subnet は NAT gateway をまだ作らないため、この route table には関連付けない。
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-public-rt"
      Role = "public"
    }
  )
}

# public subnet からインターネット宛ての通信を Internet Gateway に向ける default route。
resource "aws_route" "public_internet" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

# 各 public subnet を public route table に関連付ける。
# これにより public subnet 内のリソースは Internet Gateway への経路を持つ。
resource "aws_route_table_association" "public" {
  for_each = local.public_subnets

  subnet_id      = aws_subnet.public[each.key].id
  route_table_id = aws_route_table.public.id
}
