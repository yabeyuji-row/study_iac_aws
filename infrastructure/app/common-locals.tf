# 手動編集: 可 - 派生する名前、タグ、AZ、CIDR 計算を意図して変更する。
# ローカル値は 7-1-3 で追加する。

# 現在の AWS リージョンで利用可能なアベイラビリティゾーン一覧を取得する。
# 例: ap-northeast-1a, ap-northeast-1c, ap-northeast-1d など。
data "aws_availability_zones" "available" {
  count = length(var.availability_zone_names) == 0 ? 1 : 0

  state = "available"
}

locals {
  # リソース名の先頭に付ける共通の接頭辞。
  # 例: project_name が study-aws、environment が dev なら study-aws-dev になる。
  name_prefix = "${var.project_name}-${var.environment}"

  # すべてのアプリケーションリソースに付ける共通タグ。
  # Project/Environment/ManagedBy を必ず付け、var.tags で追加タグを受け取る。
  common_tags = merge(
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
    },
    var.tags
  )

  # AZ 名が明示されている場合はその値を使い、空の場合は AWS から取得した一覧を使う。
  # ローカル plan では availability_zone_names を指定すると AWS API 呼び出しを避けやすい。
  source_availability_zones = length(var.availability_zone_names) > 0 ? var.availability_zone_names : data.aws_availability_zones.available[0].names

  # AZ 一覧から、var.az_count で指定した数だけ先頭から取り出す。
  # 例: az_count が 2 なら、AZ のうち最初の2つを使う。
  availability_zones = slice(
    local.source_availability_zones,
    0,
    var.az_count
  )

  # パブリックサブネット用の CIDR ブロック一覧を作る。
  # cidrsubnet(var.vpc_cidr, 8, index) は VPC の CIDR をさらに細かいサブネットに分割する。
  # 例: vpc_cidr が 10.0.0.0/16 の場合、10.0.0.0/24, 10.0.1.0/24 ... のように作る。
  # cidrsubnet(prefix, newbits, netnum)
  # prefix: 元になる CIDR。例: 10.0.0.0/16
  # newbits: prefix length を何 bit 増やすか
  # netnum: その中の何番目のサブネットを取るか
  public_subnet_cidrs = [
    for index in range(var.az_count) :
    cidrsubnet(var.vpc_cidr, 8, index)
  ]

  # プライベートサブネット用の CIDR ブロック一覧を作る。
  # index + 100 にして、パブリックサブネットと重ならない離れた範囲を使う。
  # 例: vpc_cidr が 10.0.0.0/16 の場合、10.0.100.0/24, 10.0.101.0/24 ... のように作る。
  private_subnet_cidrs = [
    for index in range(var.az_count) :
    cidrsubnet(var.vpc_cidr, 8, index + 100)
  ]

  # パブリックサブネット作成用の map。
  # キーは AZ 名、値はサブネット名・AZ・CIDR をまとめた object になる。
  # aws-subnet.tf の aws_subnet.public で for_each に渡して、AZ ごとに1つずつ作成する。
  public_subnets = {
    for index, az in local.availability_zones :
    az => {
      name = "${local.name_prefix}-public-${index + 1}"
      az   = az
      cidr = local.public_subnet_cidrs[index]
    }
  }

  # プライベートサブネット作成用の map。
  # public_subnets と同じ構造で、プライベート用の名前と CIDR を持つ。
  # aws-subnet.tf の aws_subnet.private で for_each に渡して、AZ ごとに1つずつ作成する。
  private_subnets = {
    for index, az in local.availability_zones :
    az => {
      name = "${local.name_prefix}-private-${index + 1}"
      az   = az
      cidr = local.private_subnet_cidrs[index]
    }
  }
}
