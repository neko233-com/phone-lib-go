package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
)

// runCLI invokes run with captured stdout and stderr.
func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestPrettyCNLookup(t *testing.T) {
	code, stdout, stderr := runCLI("--country", "CN", "13128500205")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	for _, want := range []string{
		"号码", "13128500205",
		"E.164", "+8613128500205",
		"国际格式", "+86 131 2850 0205",
		"国内格式", "131 2850 0205",
		"中国 (CN) +86",
		"广东珠海",
		"中国联通",
		"号段", "131",
		"区号", "0756",
		"邮编", "519000",
		"手机",
		"Asia/Shanghai",
		"有效",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "无效") {
		t.Errorf("stdout reports 无效 for a valid number:\n%s", stdout)
	}
}

func TestFlagsAfterNumber(t *testing.T) {
	_, wantPretty, _ := runCLI("--country", "CN", "13128500205")

	code, gotPretty, stderr := runCLI("13128500205", "--country", "CN")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if gotPretty != wantPretty {
		t.Errorf("flags after number produced different output:\ngot  %q\nwant %q", gotPretty, wantPretty)
	}

	code, stdout, stderr := runCLI("+8613128500205", "--json")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	var info phone.Info
	if err := json.Unmarshal([]byte(stdout), &info); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", stdout, err)
	}
	if info.E164 != "+8613128500205" || info.Raw != "+8613128500205" {
		t.Errorf("decoded Info = E164 %q Raw %q, want both +8613128500205", info.E164, info.Raw)
	}
}

func TestJSONLookup(t *testing.T) {
	code, stdout, stderr := runCLI("--json", "--country", "CN", "13128500205")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var info phone.Info
	if err := json.Unmarshal([]byte(stdout), &info); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", stdout, err)
	}
	checks := []struct {
		name string
		got  any
		want any
	}{
		{"Type", info.Type, phone.TypeMobile},
		{"E164", info.E164, "+8613128500205"},
		{"ISO", info.ISO, "CN"},
		{"Country", info.Country, "中国"},
		{"Location", info.Location, "广东珠海"},
		{"Carrier", info.Carrier, "中国联通"},
		{"Prefix", info.Prefix, "131"},
		{"AreaCode", info.AreaCode, "0756"},
		{"ZipCode", info.ZipCode, "519000"},
		{"Valid", info.Valid, true},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("decoded Info %s = %v, want %v", c.name, c.got, c.want)
		}
	}

	// The wire format must encode the type as the documented string name.
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("json.Unmarshal into map error = %v", err)
	}
	if got, ok := payload["type"].(string); !ok || got != "mobile" {
		t.Errorf("JSON type = %#v, want string \"mobile\"", payload["type"])
	}
	if got := payload["area_code"]; got != "0756" {
		t.Errorf("JSON area_code = %#v, want \"0756\"", got)
	}
}

func TestCountriesCommand(t *testing.T) {
	code, stdout, stderr := runCLI("--countries")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	count := len(phone.Countries())
	if count != 245 {
		t.Errorf("len(phone.Countries()) = %d, want 245", count)
	}
	if want := fmt.Sprintf("共 %d 个国家/地区", count); !strings.Contains(stdout, want) {
		t.Errorf("stdout does not contain %q", want)
	}
	for _, want := range []string{"ISO", "区号", "中文名", "英文名", "旗帜", "CN", "+86", "中国", "🇨🇳"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q", want)
		}
	}

	// The CN row must be its own table row.
	foundCN := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(line, "CN ") {
			foundCN = true
			if !strings.Contains(line, "+86") || !strings.Contains(line, "中国") {
				t.Errorf("CN row = %q, want calling code +86 and name 中国", line)
			}
		}
	}
	if !foundCN {
		t.Errorf("--countries output has no CN row:\n%s", stdout)
	}
}

func TestVersionCommand(t *testing.T) {
	code, stdout, stderr := runCLI("--version")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if want := "phone-lib-go v" + phone.Version + "\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func TestCommandErrors(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr []string
	}{
		{
			name:       "missing number",
			args:       nil,
			wantStderr: []string{"缺少电话号码"},
		},
		{
			name:       "two numbers",
			args:       []string{"13128500205", "07561234567"},
			wantStderr: []string{"只接受一个电话号码", "收到 2 个参数"},
		},
		{
			name:       "bad country",
			args:       []string{"--country", "ZZ", "13128500205"},
			wantStderr: []string{"错误", "unknown country"},
		},
		{
			name:       "national without hint",
			args:       []string{"13128500205"},
			wantStderr: []string{"错误", "country hint"},
		},
		{
			name:       "unknown flag",
			args:       []string{"--nope"},
			wantStderr: []string{"flag provided but not defined"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(tt.args...)
			if code != 1 {
				t.Errorf("run exit code = %d, want 1", code)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Errorf("stderr is empty, want an error message")
			}
			for _, want := range tt.wantStderr {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr does not contain %q:\n%s", want, stderr)
				}
			}
		})
	}
}

func TestParsedButInvalidNumber(t *testing.T) {
	// +1 200 555 0123 parses as a US number but 200 is not an assigned area
	// code, so the CLI must print the result with 校验 无效 and exit 0.
	code, stdout, stderr := runCLI("+12005550123")
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "无效") {
		t.Errorf("stdout does not report 无效:\n%s", stdout)
	}
	if strings.Contains(stdout, "有效") {
		t.Errorf("stdout reports 有效 for an invalid number:\n%s", stdout)
	}
	if !strings.Contains(stdout, "校验") {
		t.Errorf("stdout has no 校验 line:\n%s", stdout)
	}
}
