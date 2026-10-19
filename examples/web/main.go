// Command web is a self-contained web UI demo for the phone library, styled
// after a 114 directory-enquiry page: a blue header, a search card and a
// result grid for 归属地, 运营商, 号段, 区号, 邮编, 类型, 时区 and 校验.
//
// It uses only the Go standard library and embeds all HTML, CSS and JavaScript
// in the binary, so the page loads no external assets and works offline.
//
// Usage:
//
//	go run ./examples/web
//	go run ./examples/web -addr 127.0.0.1:9090
//
// Then open http://127.0.0.1:8080 and search for a phone number.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	phone "github.com/neko233-com/phone-lib-go"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/api/lookup", lookupHandler)

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("phone-lib-go web demo listening on http://%s", *addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// indexHandler serves the single-page demo at "/" and nothing else.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.WriteString(w, indexHTML)
}

// lookupHandler serves GET /api/lookup?number=...&country=...
//
// It returns the phone.Info value as JSON on success and {"error":"..."} with a
// 4xx/5xx status on failure. A number that parses but fails the numbering plan
// is not an error: it is returned with "valid": false and HTTP 200.
func lookupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed: use GET")
		return
	}

	query := r.URL.Query()
	number := strings.TrimSpace(query.Get("number"))
	if number == "" {
		writeError(w, http.StatusBadRequest, "missing required query parameter: number")
		return
	}
	country := strings.TrimSpace(query.Get("country"))

	info, err := lookup(number, country)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// lookup calls the library, passing a country hint only when one was given.
func lookup(number, country string) (*phone.Info, error) {
	if country == "" {
		return phone.Lookup(number)
	}
	return phone.Lookup(number, phone.WithCountry(country))
}

// writeJSON writes v as an indented JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

// writeError writes {"error": msg} with the given HTTP status.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// indexHTML is the complete single-page demo. It has no external dependencies:
// all CSS and JavaScript are inline. The JavaScript builds the result with
// document.createElement / textContent, so lookup values can never inject HTML.
const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>114号码通 · phone-lib-go 演示</title>
<style>
  :root {
    --blue: #1a5fd0;
    --blue-dark: #0f3f96;
    --blue-soft: #eef4ff;
    --bg: #eef3fb;
    --ink: #1f2937;
    --muted: #6b7280;
    --line: #dbe4f3;
    --ok: #0a7d3e;
    --ok-bg: #e7f6ee;
    --bad: #c0392b;
    --bad-bg: #fdecea;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue",
      Arial, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
    background: var(--bg);
    color: var(--ink);
  }
  header {
    background: linear-gradient(135deg, var(--blue), var(--blue-dark));
    color: #fff;
    padding: 26px 20px;
    box-shadow: 0 2px 10px rgba(15, 63, 150, .18);
  }
  .wrap { max-width: 780px; margin: 0 auto; padding: 0 16px; }
  .brand { display: flex; align-items: center; gap: 14px; }
  .logo {
    width: 48px; height: 48px; flex: 0 0 48px;
    border-radius: 12px; background: #fff; color: var(--blue);
    font-weight: 800; font-size: 19px; letter-spacing: .5px;
    display: flex; align-items: center; justify-content: center;
  }
  h1 { margin: 0; font-size: 21px; letter-spacing: .5px; }
  header p { margin: 5px 0 0; font-size: 13px; opacity: .88; }
  main { padding: 22px 0 40px; }
  .card {
    background: #fff;
    border: 1px solid #e5ecf8;
    border-radius: 16px;
    box-shadow: 0 10px 30px rgba(15, 63, 150, .10);
    padding: 22px;
  }
  .search { display: flex; flex-wrap: wrap; gap: 10px; align-items: stretch; }
  .search input[type="text"] {
    flex: 1 1 260px; min-width: 0;
    padding: 12px 14px;
    border: 1px solid var(--line); border-radius: 10px;
    font-size: 16px; color: var(--ink); outline: none;
  }
  .search input[type="text"]:focus {
    border-color: var(--blue);
    box-shadow: 0 0 0 3px rgba(26, 95, 208, .15);
  }
  .search .country {
    flex: 0 0 96px; width: 96px; text-align: center;
    font-size: 15px; letter-spacing: .5px;
  }
  .search button {
    flex: 0 0 auto;
    padding: 12px 30px;
    border: 0; border-radius: 10px;
    background: var(--blue); color: #fff;
    font-size: 16px; font-weight: 600; letter-spacing: 2px;
    cursor: pointer; transition: background .15s ease;
  }
  .search button:hover { background: var(--blue-dark); }
  .search button:disabled { opacity: .6; cursor: default; }
  .hint { margin: 12px 0 0; color: var(--muted); font-size: 13px; line-height: 1.8; }
  .hint code {
    padding: 1px 6px; border-radius: 5px;
    background: #f1f5fb; color: var(--blue-dark); font-size: 12px;
  }
  .hint a { color: var(--blue); text-decoration: none; }
  .hint a:hover { text-decoration: underline; }
  .msg {
    display: none;
    margin-top: 16px; padding: 11px 14px;
    border-radius: 10px; font-size: 14px;
  }
  .msg.error { display: block; background: var(--bad-bg); color: var(--bad); border: 1px solid #f5c6c0; }
  .msg.loading { display: block; background: var(--blue-soft); color: var(--blue-dark); border: 1px solid #cfe0ff; }
  .result { margin-top: 20px; }
  .result-head {
    display: flex; justify-content: space-between; align-items: flex-start;
    gap: 12px; padding-bottom: 14px; margin-bottom: 16px;
    border-bottom: 1px dashed var(--line);
  }
  .phone {
    color: var(--blue-dark); font-size: 26px; font-weight: 700;
    letter-spacing: .5px; word-break: break-all;
  }
  .phone-sub { margin-top: 4px; color: var(--muted); font-size: 13px; }
  .badge {
    flex: 0 0 auto; padding: 5px 14px; border-radius: 999px;
    font-size: 13px; font-weight: 600; white-space: nowrap;
  }
  .badge.valid { background: var(--ok-bg); color: var(--ok); }
  .badge.invalid { background: var(--bad-bg); color: var(--bad); }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(148px, 1fr));
    gap: 10px 14px;
  }
  .item {
    padding: 10px 12px;
    border: 1px solid #e8eff9; border-radius: 10px;
    background: #f7faff;
  }
  .item .k { margin-bottom: 4px; color: var(--muted); font-size: 12px; }
  .item .v { font-size: 15px; font-weight: 600; word-break: break-all; }
  .item.wide { grid-column: 1 / -1; }
  footer {
    padding: 0 16px 30px;
    color: var(--muted); font-size: 12px;
    text-align: center; line-height: 1.9;
  }
  @media (max-width: 520px) {
    h1 { font-size: 18px; }
    .search .country { flex: 0 0 80px; width: 80px; }
    .phone { font-size: 22px; }
  }
</style>
</head>
<body>
<header>
  <div class="wrap brand">
    <div class="logo">114</div>
    <div>
      <h1>114号码通 · 号码归属地查询</h1>
      <p>phone-lib-go 在线演示 · 数据内嵌二进制，全部本地计算，不访问外部服务</p>
    </div>
  </div>
</header>

<main class="wrap">
  <section class="card">
    <form class="search" id="lookup-form" autocomplete="off">
      <input type="text" id="number" name="number" value="13128500205"
             placeholder="输入电话号码，如 13128500205 或 +8613128500205"
             aria-label="电话号码" spellcheck="false">
      <input type="text" id="country" name="country" class="country" value="CN"
             placeholder="国家提示" aria-label="国家提示" spellcheck="false">
      <button type="submit" id="search-btn">查询</button>
    </form>
    <p class="hint">
      国际格式（<code>+</code> 或 <code>00</code> 开头）时国家提示可留空；国内格式需要填写
      <code>CN</code>、<code>86</code> 或 <code>中国</code> 等提示。
      示例：<a href="?number=13128500205&amp;country=CN">13128500205</a> ·
      <a href="?number=%2B8613128500205&amp;country=">+8613128500205</a> ·
      <a href="?number=2025550123&amp;country=US">2025550123（美国）</a>
    </p>

    <div class="msg" id="message" role="status" aria-live="polite"></div>

    <section class="result" id="result" hidden>
      <div class="result-head">
        <div>
          <div class="phone" id="r-e164"></div>
          <div class="phone-sub" id="r-national"></div>
        </div>
        <span class="badge" id="r-valid"></span>
      </div>
      <div class="grid">
        <div class="item"><div class="k">归属地</div><div class="v" id="r-location"></div></div>
        <div class="item"><div class="k">运营商</div><div class="v" id="r-carrier"></div></div>
        <div class="item"><div class="k">号段</div><div class="v" id="r-prefix"></div></div>
        <div class="item"><div class="k">区号</div><div class="v" id="r-area"></div></div>
        <div class="item"><div class="k">邮编</div><div class="v" id="r-zip"></div></div>
        <div class="item"><div class="k">类型</div><div class="v" id="r-type"></div></div>
        <div class="item"><div class="k">时区</div><div class="v" id="r-timezones"></div></div>
        <div class="item"><div class="k">国家/地区</div><div class="v" id="r-country"></div></div>
      </div>
    </section>
  </section>
</main>

<footer>
  phone-lib-go 演示 · 全球数据来自 Google libphonenumber 元数据，中国大陆号段来自
  iBugCode/phone2region · 仅供学习参考，不作为号码真实性依据
</footer>

<script>
(function () {
  'use strict';

  var TYPE_ZH = {
    unknown: '未知',
    fixed_line: '固定电话',
    mobile: '手机',
    fixed_line_or_mobile: '固定电话或手机',
    toll_free: '免费电话',
    premium_rate: '付费电话',
    shared_cost: '共享成本号码',
    voip: '网络电话',
    personal_number: '个人号码',
    pager: '寻呼机',
    uan: '通用接入号码',
    voicemail: '语音信箱'
  };

  var form = document.getElementById('lookup-form');
  var numberInput = document.getElementById('number');
  var countryInput = document.getElementById('country');
  var searchButton = document.getElementById('search-btn');
  var message = document.getElementById('message');
  var result = document.getElementById('result');

  function setText(id, value) {
    var el = document.getElementById(id);
    if (el) { el.textContent = value; }
  }

  function dash(value) {
    if (value === undefined || value === null || value === '') { return '-'; }
    return String(value);
  }

  function carrierText(info) {
    var zh = info.carrier || '';
    var en = info.carrier_en || '';
    if (!zh && !en) { return '-'; }
    if (!zh) { return en; }
    if (!en || en === zh) { return zh; }
    return zh + ' (' + en + ')';
  }

  function countryText(info) {
    var name = info.country || info.iso || '-';
    var parts = [];
    if (info.iso) { parts.push(info.iso); }
    if (info.country_code) { parts.push('+' + info.country_code); }
    var line = parts.length ? name + ' (' + parts.join(' ') + ')' : name;
    if (info.flag) { line += ' ' + info.flag; }
    return line;
  }

  function timezonesText(info) {
    if (!Array.isArray(info.timezones) || info.timezones.length === 0) { return '-'; }
    return info.timezones.join(', ');
  }

  function showMessage(text, kind) {
    message.textContent = text;
    message.className = 'msg ' + kind;
  }

  function clearMessage() {
    message.textContent = '';
    message.className = 'msg';
  }

  function render(info) {
    setText('r-e164', dash(info.e164 || info.raw));
    setText('r-national', info.national ? '国内格式 ' + info.national : '');
    setText('r-location', dash(info.location));
    setText('r-carrier', carrierText(info));
    setText('r-prefix', dash(info.prefix));
    setText('r-area', dash(info.area_code));
    setText('r-zip', dash(info.zip_code));
    setText('r-type', TYPE_ZH[info.type] || dash(info.type));
    setText('r-timezones', timezonesText(info));
    setText('r-country', countryText(info));

    var badge = document.getElementById('r-valid');
    badge.textContent = info.valid ? '有效' : '无效';
    badge.className = 'badge ' + (info.valid ? 'valid' : 'invalid');

    result.hidden = false;
  }

  function lookup() {
    var number = numberInput.value.trim();
    if (!number) {
      result.hidden = true;
      showMessage('请输入电话号码', 'error');
      return;
    }
    var country = countryInput.value.trim();
    var url = '/api/lookup?number=' + encodeURIComponent(number) +
              '&country=' + encodeURIComponent(country);

    searchButton.disabled = true;
    showMessage('查询中…', 'loading');

    fetch(url, { headers: { 'Accept': 'application/json' } })
      .then(function (response) {
        return response.json().then(function (body) {
          return { ok: response.ok, status: response.status, body: body };
        });
      })
      .then(function (res) {
        if (!res.ok) {
          var reason = (res.body && res.body.error) ? res.body.error : ('HTTP ' + res.status);
          throw new Error(reason);
        }
        clearMessage();
        render(res.body);
      })
      .catch(function (err) {
        result.hidden = true;
        showMessage(err && err.message ? err.message : '查询失败，请稍后重试', 'error');
      })
      .then(function () {
        searchButton.disabled = false;
      });
  }

  form.addEventListener('submit', function (event) {
    event.preventDefault();
    lookup();
  });

  // Prefill from the URL (?number=...&country=...) and query immediately so a
  // shared link opens directly on its result.
  var params = new URLSearchParams(window.location.search);
  if (params.has('number')) { numberInput.value = params.get('number'); }
  if (params.has('country')) { countryInput.value = params.get('country'); }
  lookup();
})();
</script>
</body>
</html>
`
