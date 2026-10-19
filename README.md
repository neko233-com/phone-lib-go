# phone-lib-go

**114 风格的电话号码解析、校验与归属地查询库（Go）**

[![Go Reference](https://pkg.go.dev/badge/github.com/neko233-com/phone-lib-go.svg)](https://pkg.go.dev/github.com/neko233-com/phone-lib-go)
[![CI](https://github.com/neko233-com/phone-lib-go/actions/workflows/ci.yml/badge.svg)](https://github.com/neko233-com/phone-lib-go/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`phone-lib-go` 用一行代码解析任意格式的电话号码，返回标准格式、国家/地区、归属地（省/市）、运营商、号段、区号、邮编、号码类型与时区。全球数据来自 Google libphonenumber 元数据，中国大陆手机号还能定位到城市级别。

```go
info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
// 号码 13128500205 · 归属地 广东珠海 · 运营商 中国联通 · 区号 0756 · 类型 手机
```

## 特性

- 🌍 **全球号码支持**：覆盖 libphonenumber 的全部国家与地区，支持解析、校验、E.164 / 国际 / 国内 / RFC 3966 格式化。
- 🇨🇳 **中国大陆号段数据**：内置号段库，手机号可解析到省份、城市、区号、邮编与运营商。
- 🔎 **灵活的国家提示**：ISO 代码（`CN`）、国际区号（`86`、`+86`）、中英文名称（`中国`、`United States`）、常见别名（`香港`、`USA`）均可。
- 🏷️ **号码类型识别**：手机、固定电话、免费电话、付费电话、VoIP、寻呼机等 12 种类型。
- 🕐 **时区与归属地**：返回 IANA 时区与中英文归属地；中国大陆为「省+市」，部分国家英文归属地可到城市。
- 🧾 **JSON 友好**：`Info` 所有字段带 JSON tag，`type` 序列化为 `"mobile"` 等字符串，可直接对外提供 API。
- 🖥️ **自带 CLI**：`phone` 命令提供 114 风格中文输出，支持 `--json`、`--countries`、`--version`。
- 📦 **纯 Go、离线运行**：数据嵌入二进制，不做运行时联网查询；直接依赖只有 `nyaruka/phonenumbers` 与 `x/text`。

## 安装

```bash
go get github.com/neko233-com/phone-lib-go
```

要求 Go 1.25 或更高版本。

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	phone "github.com/neko233-com/phone-lib-go"
)

func main() {
	info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info.Location, info.Carrier, info.Prefix, info.AreaCode)
	// 广东珠海 中国联通 131 0756
}
```

以 `+` 或 `00` 开头的国际格式不需要国家提示：

```go
info, err := phone.Lookup("+442083661177")
if err != nil {
	log.Fatal(err)
}
fmt.Println(info.E164, info.Location, info.AreaCode)
// +442083661177 英国 020
```

能解析但不符合号码规划规则的输入会返回 `info.Valid == false` 且 `err == nil`；只有无法解析的输入，或无法识别的国家提示才返回错误：

```go
info, err := phone.Lookup("12345", phone.WithCountry("CN"))
if err != nil {
	log.Fatal(err)
}
if !info.Valid {
	fmt.Println("无效号码") // 输出: 无效号码
}
```

## 输出示例

命令行 `phone --country CN 13128500205` 的完整输出（字段值与库返回值一致）：

```text
号码         13128500205
E.164        +8613128500205
国际格式     +86 131 2850 0205
国内格式     131 2850 0205
国家/地区    中国 (CN) +86 🇨🇳
归属地       广东珠海
运营商       中国联通 (China Unicom)
号段         131
区号         0756
邮编         519000
类型         手机
时区         Asia/Shanghai
校验         有效
```

`phone --json --country CN 13128500205` 输出（JSON 字段与 Go 结构体一一对应）：

```json
{
  "raw": "13128500205",
  "e164": "+8613128500205",
  "international": "+86 131 2850 0205",
  "national": "131 2850 0205",
  "rfc3966": "tel:+86-131-2850-0205",
  "national_number": "13128500205",
  "country_code": 86,
  "iso": "CN",
  "country": "中国",
  "country_en": "China",
  "flag": "🇨🇳",
  "valid": true,
  "possible": true,
  "type": "mobile",
  "carrier": "中国联通",
  "carrier_en": "China Unicom",
  "location": "广东珠海",
  "location_en": "Zhuhai, Guangdong",
  "province": "广东",
  "city": "珠海",
  "prefix": "131",
  "area_code": "0756",
  "zip_code": "519000",
  "timezones": [
    "Asia/Shanghai"
  ]
}
```

## 命令行工具

安装：

```bash
go install github.com/neko233-com/phone-lib-go/cmd/phone@latest
```

在仓库内也可以直接运行 `go run ./cmd/phone`，或用 `make build` 编译出 `./phone`（Windows 上为 `./phone.exe`）。

```bash
# 国内格式：需要 --country 提示
go run ./cmd/phone --country CN 13128500205

# 国际格式：无需提示，直接以 + 或 00 开头
go run ./cmd/phone +8613128500205

# JSON 输出
go run ./cmd/phone --json --country CN 13128500205

# 列出所有支持的国家/地区
go run ./cmd/phone --countries

# 版本信息
go run ./cmd/phone --version
# phone-lib-go v0.1.0
```

选项可以写在号码前面或后面，两种写法等价：

```bash
phone +8613128500205 --json
phone --json --country CN 13128500205
```

| 选项 | 说明 |
| --- | --- |
| `-country <hint>` | 国家/地区提示，供国内格式号码使用：ISO 代码 `CN`、区号 `86`/`+86`、中文名 `中国`、英文名 `United States`、别名 `香港`/`USA` |
| `-json` | 以缩进 JSON 输出完整 `Info` |
| `-countries` | 输出国家/地区对照表（ISO、区号、中文名、英文名、旗帜），忽略号码参数 |
| `-version` | 输出 `phone-lib-go v0.1.0` |
| `-h`, `--help` | 输出用法说明 |

退出码：`0` 表示查询成功（解析成功但号码无效也是 `0`，输出中「校验」为 `无效`）；参数错误、无法解析或国家提示非法返回 `1`，错误信息写入 stderr。

`phone --countries` 共输出 245 个国家/地区，节选：

```text
ISO  区号  中文名                 英文名                          旗帜
----------------------------------------------------------------------
AC   +247  阿森松岛               Ascension Island                🇦🇨
AD   +376  安道尔                 Andorra                         🇦🇩
AE   +971  阿拉伯联合酋长国       United Arab Emirates            🇦🇪
...
ZW   +263  津巴布韦               Zimbabwe                        🇿🇼

共 245 个国家/地区
```

## API 参考

### 函数

| 函数 | 说明 |
| --- | --- |
| `func Lookup(raw string, opts ...Option) (*Info, error)` | 解析并补全号码信息；国内格式需要 `WithCountry` |
| `func Parse(raw, country string) (*Info, error)` | `Lookup(raw, WithCountry(country))` 的简写 |
| `func MustLookup(raw string, opts ...Option) *Info` | 同 `Lookup`，出错时 panic |
| `func MustParse(raw, country string) *Info` | 同 `Parse`，出错时 panic |
| `func WithCountry(country string) Option` | 国家提示：ISO 代码、区号（`86`/`+86`）、中英文名、常见别名 |
| `func Countries() []Country` | 全部支持的国家/地区，按 ISO 代码排序 |
| `func CountryByISO(iso string) (Country, bool)` | 按 ISO 3166-1 alpha-2 查询国家/地区 |
| `func CountryByCallingCode(code int) (Country, bool)` | 按国际区号查询主国家，如 `86` → 中国、`1` → 美国 |
| `func ChinaDataVersion() string` | 中国大陆号段数据版本，如 `"2025-02"` |

### Info 字段

| 字段 | JSON | 说明 |
| --- | --- | --- |
| `Raw` | `raw` | 原始输入 |
| `E164` | `e164` | E.164 格式，如 `+8613128500205` |
| `International` | `international` | 国际格式，如 `+86 131 2850 0205` |
| `National` | `national` | 国内格式，如 `131 2850 0205` |
| `RFC3966` | `rfc3966` | RFC 3966 / `tel:` URI，如 `tel:+86-131-2850-0205` |
| `NationalNumber` | `national_number` | 国内有效号码（不含国际区号与国内前缀） |
| `CountryCode` | `country_code` | 国际区号，如 `86` |
| `ISO` | `iso` | ISO 3166-1 alpha-2 代码，如 `CN` |
| `Country` | `country` | 国家/地区中文名，如 `中国` |
| `CountryEn` | `country_en` | 国家/地区英文名，如 `China` |
| `Flag` | `flag` | Emoji 旗帜，如 `🇨🇳` |
| `Valid` | `valid` | 是否符合号码规划规则 |
| `Possible` | `possible` | 长度是否可能合理（不校验完整规则） |
| `Type` | `type` | 号码类型，JSON 中为 `"mobile"` 等字符串 |
| `Carrier` | `carrier` | 运营商中文名，如 `中国联通`；无中文数据时回退英文名 |
| `CarrierEn` | `carrier_en` | 运营商英文名，如 `China Unicom` |
| `Location` | `location` | 归属地中文名；中国大陆为「省+市」，如 `广东珠海` |
| `LocationEn` | `location_en` | 归属地英文名，如 `Zhuhai, Guangdong` |
| `Province` | `province` | 省份（中国大陆号码），如 `广东` |
| `City` | `city` | 城市（中国大陆号码），如 `珠海` |
| `Prefix` | `prefix` | 号段，如 `131`；中国大陆固话为区号 |
| `AreaCode` | `area_code` | 区号，如 `0756`、`020`、`03`、`202` |
| `ZipCode` | `zip_code` | 邮编，目前中国大陆手机号提供，如 `519000` |
| `Timezones` | `timezones` | IANA 时区列表，如 `["Asia/Shanghai"]` |
| `Extension` | `extension` | 分机号（输入中包含时） |

方法：`(*Info).IsMobile() bool` 判断是否手机（含固话/手机二选一的号段），`(*Info).String() string` 输出一行摘要。

### Type 类型

| 常量 | `String()` | `Chinese()` |
| --- | --- | --- |
| `TypeUnknown` | `unknown` | 未知 |
| `TypeFixedLine` | `fixed_line` | 固定电话 |
| `TypeMobile` | `mobile` | 手机 |
| `TypeFixedLineOrMobile` | `fixed_line_or_mobile` | 固定电话或手机 |
| `TypeTollFree` | `toll_free` | 免费电话 |
| `TypePremiumRate` | `premium_rate` | 付费电话 |
| `TypeSharedCost` | `shared_cost` | 共享成本号码 |
| `TypeVoIP` | `voip` | 网络电话 |
| `TypePersonalNumber` | `personal_number` | 个人号码 |
| `TypePager` | `pager` | 寻呼机 |
| `TypeUAN` | `uan` | 通用接入号码 |
| `TypeVoicemail` | `voicemail` | 语音信箱 |

### Country

| 字段 | JSON | 说明 |
| --- | --- | --- |
| `ISO` | `iso` | ISO 3166-1 alpha-2 代码 |
| `CallingCode` | `calling_code` | 国际区号 |
| `Name` | `name` | 中文名 |
| `NameEn` | `name_en` | 英文名 |
| `Flag` | `flag` | Emoji 旗帜 |

### 错误

| 错误 | 含义 |
| --- | --- |
| `ErrCountryRequired` | 国内格式号码缺少国家提示 |
| `ErrUnknownCountry` | 无法识别的国家提示 |
| `ErrEmptyNumber` | 输入为空 |

## 数据来源与版本

- **全球号码元数据**：Google libphonenumber metadata **v9.0.40**，通过 `github.com/nyaruka/phonenumbers/v2` **v2.0.14** 引入（Apache-2.0）。提供解析、校验、格式化、号码类型、归属地描述、运营商与 IANA 时区数据。
- **中国大陆号段**：iBugCode/phone2region 的 `phone.dat`，数据版本 **2025-02**（MIT），以 `go:embed` 嵌入二进制。提供手机号的省份、城市、区号、邮编与运营商。
- **国家/地区名称**：`golang.org/x/text`（CLDR），生成中英文国家名称与旗帜。
- 运行时可通过 `phone.ChinaDataVersion()` 获取内置的中国号段数据版本；第三方声明见 [NOTICE.md](NOTICE.md)。

## 覆盖范围与限制

- **全球**：国家/地区、号码类型、E.164/国际/国内格式、运营商、时区均可获得；中文归属地通常只到国家/地区级，`LocationEn` 对部分国家可到城市级（如美国的 `Washington D.C.`）。
- **中国大陆**：手机号可解析到省份、城市、区号与邮编（来自内置号段库）；固定电话的归属地与区号由 libphonenumber 元数据推导。手机号未命中号段时，`Province`、`City`、`AreaCode`、`ZipCode` 为空字符串；固定电话没有号段记录，但会根据区号推导 `Province`/`City`。
- **运营商为推断值**：根据号段/号段元数据推断，携号转网（号码携带）后可能与实际不符。
- **数据为离线快照**：不是实时数据，新增号段或号码规划调整需要更新依赖版本或号段库。
- **不联网校验**：只做格式与号段校验，不验证号码是否真实存在、是否停机。

## English

`phone-lib-go` is a 114-style phone number lookup library for Go. It parses,
validates and formats phone numbers from every region supported by Google's
libphonenumber metadata, returning country/region, location, carrier, 号段
(prefix), area code, postal code, number type, timezones and a JSON-friendly
`Info` value. Mainland Chinese mobile numbers resolve to province/city level
via an embedded 号段 database.

```bash
go get github.com/neko233-com/phone-lib-go
```

```go
info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
if err != nil {
	log.Fatal(err)
}
fmt.Println(info.E164, info.Location, info.Carrier, info.AreaCode)
// +8613128500205 广东珠海 中国联通 0756
```

International-format numbers (`+` or `00` prefix) need no country hint. A
successfully parsed but invalid number returns `Info.Valid == false` with a nil
error. A CLI is included:

```bash
phone --country CN 13128500205
phone +8613128500205 --json
phone --countries
```

Requires Go 1.25+. MIT licensed; see [LICENSE](LICENSE) and
[NOTICE.md](NOTICE.md) for third-party attributions.

## 开发

```bash
go test ./...        # 单元测试
go test -race ./...  # 竞态检测
go vet ./...         # 静态检查
go run ./cmd/phone --country CN 13128500205
```

Makefile 快捷命令：`make build`、`make test`、`make race`、`make vet`、`make fmt`、`make run-demo`。

CI 使用 GitHub Actions，在 Go 1.25.x、1.26.x、1.27.x 上执行 gofmt 检查、`go vet ./...`、`go build ./...` 与 `go test -race ./...`。

## 许可证与致谢

本项目使用 [MIT License](LICENSE)，Copyright (c) 2026 neko233-com。

- [nyaruka/phonenumbers](https://github.com/nyaruka/phonenumbers)（Apache-2.0）
- [Google libphonenumber](https://github.com/google/libphonenumber)（Apache-2.0），含 metadata / geocoding / carrier 数据
- [iBugCode/phone2region](https://github.com/iBugCode/phone2region)（MIT），中国大陆号段数据，数据版本 2025-02

感谢以上项目及其维护者。完整第三方声明见 [NOTICE.md](NOTICE.md)。
