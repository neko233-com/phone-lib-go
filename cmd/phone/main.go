// Command phone is a small 114-style phone number lookup CLI built on
// github.com/neko233-com/phone-lib-go.
//
// Usage:
//
//	phone --country CN 13128500205
//	phone +8613128500205
//	phone --json --country CN 13128500205
//	phone --countries
//	phone --version
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	phone "github.com/neko233-com/phone-lib-go"
)

// version is reported by -version and kept in sync with the library version.
const version = phone.Version

// labelColumn is the display column (counting wide runes as two columns) at
// which pretty output values start.
const labelColumn = 13

const usage = `phone - 电话号码解析与归属地查询 (phone-lib-go)

用法:
  phone [选项] <电话号码>

选项:
  -country string
        国家/地区提示，供国内格式号码使用 (ISO 代码如 CN、区号如 86、中文名如 中国)
  -json
        以 JSON 格式输出
  -countries
        列出所有支持的国家/地区
  -version
        显示版本信息

示例:
  phone --country CN 13128500205
  phone +8613128500205
  phone --json --country CN 13128500205
  phone --countries
  phone --version
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI and returns the process exit code. It is split out from
// main so the command can be exercised without exiting the process.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("phone", flag.ContinueOnError)
	fs.SetOutput(stderr)

	country := fs.String("country", "", "国家/地区提示，供国内格式号码使用")
	asJSON := fs.Bool("json", false, "以 JSON 格式输出")
	listCountries := fs.Bool("countries", false, "列出所有支持的国家/地区")
	showVersion := fs.Bool("version", false, "显示版本信息")
	fs.Usage = func() { fmt.Fprint(stderr, usage) }

	// The standard flag package stops at the first non-flag argument, but users
	// naturally write "phone +8613128500205 --json". Reorder the arguments so
	// every flag is parsed regardless of where it appears on the command line.
	if err := fs.Parse(reorderArgs(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if *showVersion {
		fmt.Fprintf(stdout, "phone-lib-go v%s\n", version)
		return 0
	}

	if *listCountries {
		printCountries(stdout, phone.Countries())
		return 0
	}

	switch fs.NArg() {
	case 0:
		fmt.Fprintln(stderr, "错误: 缺少电话号码")
		fmt.Fprint(stderr, usage)
		return 1
	case 1:
		// ok
	default:
		fmt.Fprintf(stderr, "错误: 只接受一个电话号码，收到 %d 个参数: %s\n",
			fs.NArg(), strings.Join(fs.Args(), " "))
		fmt.Fprint(stderr, usage)
		return 1
	}

	opts := make([]phone.Option, 0, 1)
	if *country != "" {
		opts = append(opts, phone.WithCountry(*country))
	}
	info, err := phone.Lookup(fs.Arg(0), opts...)
	if err != nil {
		fmt.Fprintf(stderr, "错误: %v\n", err)
		return 1
	}

	if *asJSON {
		encoded, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			fmt.Fprintf(stderr, "错误: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(encoded))
		return 0
	}

	printInfo(stdout, info)
	return 0
}

// reorderArgs moves flags before operands so that flag.Parse accepts options
// written after the phone number. A literal "--" still ends flag parsing; any
// flags after it are treated as operands.
func reorderArgs(args []string) []string {
	flags := make([]string, 0, len(args))
	operands := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			operands = append(operands, args[i+1:]...)
			break
		}
		if len(arg) > 1 && arg[0] == '-' {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && flagTakesValue(arg) && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		operands = append(operands, arg)
	}
	return append(flags, operands...)
}

// flagTakesValue reports whether the named flag consumes the following
// argument. All other flags of the command are boolean.
func flagTakesValue(arg string) bool {
	return strings.TrimLeft(arg, "-") == "country"
}

// printInfo writes the pretty Chinese lookup result. A parsed-but-invalid
// number is still printed; the 校验 line reports 无效 and the exit code stays 0.
func printInfo(w io.Writer, info *phone.Info) {
	field := func(label, value string) {
		if value == "" {
			value = "-"
		}
		fmt.Fprintf(w, "%s%s\n", padLabel(label), value)
	}

	field("号码", info.Raw)
	field("E.164", info.E164)
	field("国际格式", info.International)
	field("国内格式", info.National)
	field("国家/地区", countryLine(info))
	field("归属地", info.Location)
	field("运营商", carrierLine(info))
	field("号段", info.Prefix)
	field("区号", info.AreaCode)
	field("邮编", info.ZipCode)
	field("类型", info.Type.Chinese())
	field("时区", strings.Join(info.Timezones, ", "))
	field("校验", validityLine(info.Valid))
	if info.Extension != "" {
		field("分机", info.Extension)
	}
}

// countryLine renders 中国 (CN) +86 🇨🇳.
func countryLine(info *phone.Info) string {
	name := info.Country
	if name == "" {
		name = info.ISO
	}
	if info.ISO == "" {
		return name
	}
	line := fmt.Sprintf("%s (%s) +%d", name, info.ISO, info.CountryCode)
	if info.Flag != "" {
		line += " " + info.Flag
	}
	return line
}

// carrierLine renders 中国联通 (China Unicom), falling back to whichever name
// is available.
func carrierLine(info *phone.Info) string {
	switch {
	case info.Carrier == "":
		return info.CarrierEn
	case info.CarrierEn == "" || info.CarrierEn == info.Carrier:
		return info.Carrier
	default:
		return fmt.Sprintf("%s (%s)", info.Carrier, info.CarrierEn)
	}
}

func validityLine(valid bool) string {
	if valid {
		return "有效"
	}
	return "无效"
}

// printCountries writes the supported country/region table. It is a compact
// table aligned by display width so that Chinese names line up.
func printCountries(w io.Writer, countries []phone.Country) {
	header := []string{"ISO", "区号", "中文名", "英文名", "旗帜"}
	rows := make([][]string, 0, len(countries)+1)
	rows = append(rows, header)
	for _, c := range countries {
		rows = append(rows, []string{
			c.ISO,
			"+" + strconv.Itoa(c.CallingCode),
			c.Name,
			c.NameEn,
			c.Flag,
		})
	}

	widths := make([]int, len(header))
	for _, row := range rows {
		for i, cell := range row {
			if n := displayWidth(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}

	for i, row := range rows {
		var b strings.Builder
		for j, cell := range row {
			b.WriteString(cell)
			if j < len(row)-1 {
				b.WriteString(strings.Repeat(" ", widths[j]-displayWidth(cell)+2))
			}
		}
		fmt.Fprintln(w, strings.TrimRight(b.String(), " "))
		if i == 0 {
			var sep strings.Builder
			for j, width := range widths {
				if j == 0 {
					sep.WriteString(strings.Repeat("-", width))
				} else {
					sep.WriteString(strings.Repeat("-", width+2))
				}
			}
			fmt.Fprintln(w, sep.String())
		}
	}
	fmt.Fprintf(w, "\n共 %d 个国家/地区\n", len(countries))
}

// padLabel pads label with spaces up to labelColumn display columns.
func padLabel(label string) string {
	n := labelColumn - displayWidth(label)
	if n < 1 {
		n = 1
	}
	return label + strings.Repeat(" ", n)
}

// displayWidth returns the number of terminal columns s occupies, counting
// East Asian wide runes and emoji as two columns.
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeWidth(r)
	}
	return width
}

func runeWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 0x20, r >= 0x7f && r < 0xa0:
		return 0
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r == 0x2329, r == 0x232A,
		r >= 0x2E80 && r <= 0xA4CF, // CJK ... Yi
		r >= 0xAC00 && r <= 0xD7A3, // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF, // CJK compatibility ideographs
		r >= 0xFE10 && r <= 0xFE19, // vertical forms
		r >= 0xFE30 && r <= 0xFE6F, // CJK compatibility forms
		r >= 0xFF00 && r <= 0xFF60, // fullwidth forms
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF, // emoji
		r >= 0x20000 && r <= 0x3FFFD: // CJK extensions
		return 2
	}
	return 1
}
