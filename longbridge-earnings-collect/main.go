package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const timeout = 60 * time.Second

var numRe = regexp.MustCompile(`^-?[0-9]+\.[0-9]+$`)

func die(msg string, code int) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

func fetch(outDir, name string, args []string) {
	cmdArgs := append(args, "--format", "json")
	cmd := exec.Command("longbridge", cmdArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		errMsg := "empty response"
		if err != nil {
			errMsg = err.Error()
		}
		os.WriteFile(filepath.Join(outDir, name+".err"), []byte(errMsg), 0644)
		return
	}
	os.WriteFile(filepath.Join(outDir, name+".json"), output, 0644)
}

// ── JSON trimming helpers ─────────────────────────────────────────────

func slim(node interface{}) interface{} {
	switch v := node.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{})
		for k, val := range v {
			out[k] = slim(val)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, val := range v {
			out[i] = slim(val)
		}
		return out
	case string:
		if numRe.MatchString(v) {
			n, err := strconv.ParseFloat(v, 64)
			if err == nil {
				if n > 1_000_000 || n < -1_000_000 {
					return fmt.Sprintf("%d", int64(n))
				}
				r := float64(int64(n*100)) / 100
				if r == float64(int64(r)) {
					return fmt.Sprintf("%d", int64(r))
				}
				return fmt.Sprintf("%.2f", r)
			}
		}
		return v
	default:
		return v
	}
}

func dropEmpty(node interface{}) interface{} {
	switch v := node.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{})
		for k, val := range v {
			if val == "" || val == nil {
				continue
			}
			out[k] = dropEmpty(val)
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(v))
		for _, val := range v {
			out = append(out, dropEmpty(val))
		}
		return out
	default:
		return v
	}
}

func findObjects(node interface{}, key string) []map[string]interface{} {
	var found []map[string]interface{}
	switch v := node.(type) {
	case map[string]interface{}:
		if _, ok := v[key]; ok {
			found = append(found, v)
		}
		for _, val := range v {
			found = append(found, findObjects(val, key)...)
		}
	case []interface{}:
		for _, val := range v {
			found = append(found, findObjects(val, key)...)
		}
	}
	return found
}

func pick(obj map[string]interface{}, keys ...string) map[string]interface{} {
	out := make(map[string]interface{})
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			out[k] = v
		}
	}
	return out
}

// ── Per-section trim filters ─────────────────────────────────────────

func trimStatement(data map[string]interface{}) interface{} {
	var out []interface{}
	list, _ := data["list"].(map[string]interface{})
	for _, kind := range list {
		kindMap, ok := kind.(map[string]interface{})
		if !ok {
			continue
		}
		indicators, _ := kindMap["indicators"].([]interface{})
		for _, ind := range indicators {
			indMap, _ := ind.(map[string]interface{})
			accounts, _ := indMap["accounts"].([]interface{})
			var acctsOut []interface{}
			for _, a := range accounts {
				aMap, _ := a.(map[string]interface{})
				values, _ := aMap["values"].([]interface{})
				if len(values) > 8 {
					values = values[:8]
				}
				var valsOut []interface{}
				for _, v := range values {
					vMap, _ := v.(map[string]interface{})
					valsOut = append(valsOut, pick(vMap, "period", "value", "yoy"))
				}
				acctsOut = append(acctsOut, map[string]interface{}{
					"name":   aMap["name"],
					"field":  aMap["field"],
					"values": valsOut,
				})
			}
			out = append(out, map[string]interface{}{
				"title":    indMap["title"],
				"accounts": acctsOut,
			})
		}
	}
	return out
}

func trimConsensus(data map[string]interface{}) interface{} {
	list, _ := data["list"].([]interface{})
	if len(list) > 6 {
		list = list[:6]
	}
	var periods []interface{}
	for _, p := range list {
		pMap, _ := p.(map[string]interface{})
		details, _ := pMap["details"].([]interface{})
		var detsOut []interface{}
		for _, d := range details {
			dMap, _ := d.(map[string]interface{})
			detsOut = append(detsOut, pick(dMap, "key", "name", "estimate", "actual", "comp"))
		}
		periods = append(periods, map[string]interface{}{
			"fiscal_year":   pMap["fiscal_year"],
			"fiscal_period": pMap["fiscal_period"],
			"period_text":   pMap["period_text"],
			"details":       detsOut,
		})
	}
	return map[string]interface{}{
		"currency":       data["currency"],
		"current_period": data["current_period"],
		"periods":        periods,
	}
}

func trimForecastEPS(data map[string]interface{}) interface{} {
	items, _ := data["items"].([]interface{})
	if len(items) > 3 {
		items = items[len(items)-3:]
	}
	var out []interface{}
	for _, i := range items {
		iMap, _ := i.(map[string]interface{})
		out = append(out, map[string]interface{}{
			"mean":   iMap["forecast_eps_mean"],
			"median": iMap["forecast_eps_median"],
			"high":   iMap["forecast_eps_highest"],
			"low":    iMap["forecast_eps_lowest"],
		})
	}
	return out
}

func trimNews(data map[string]interface{}) interface{} {
	items := findObjects(data, "title")
	if len(items) > 10 {
		items = items[:10]
	}
	var out []interface{}
	for _, o := range items {
		out = append(out, pick(o, "id", "title", "published_at"))
	}
	return out
}

func trimKline(data map[string]interface{}) interface{} {
	candles := findObjects(data, "close")
	if len(candles) > 20 {
		candles = candles[len(candles)-20:]
	}
	var highVals, lowVals []float64
	var recent []interface{}
	for _, c := range candles {
		if h, ok := getFloat(c, "high"); ok {
			highVals = append(highVals, h)
		}
		if l, ok := getFloat(c, "low"); ok {
			lowVals = append(lowVals, l)
		}
		recent = append(recent, map[string]interface{}{
			"d": fmt.Sprintf("%v", c["timestamp"]),
			"c": c["close"],
		})
	}
	high250 := 0.0
	low250 := 0.0
	if len(highVals) > 0 {
		high250 = maxFloat(highVals)
	}
	if len(lowVals) > 0 {
		low250 = minFloat(lowVals)
	}
	return map[string]interface{}{
		"recent":   recent,
		"high_250d": high250,
		"low_250d":  low250,
	}
}

func trimFilings(data map[string]interface{}) interface{} {
	items := findObjects(data, "title")
	items = append(items, findObjects(data, "name")...)
	if len(items) > 10 {
		items = items[:10]
	}
	var out []interface{}
	for _, o := range items {
		date := o["published_at"]
		if date == nil {
			date = o["date"]
		}
		if date == nil {
			date = o["filed_at"]
		}
		out = append(out, map[string]interface{}{
			"id":    o["id"],
			"title": firstNonNil(o["title"], o["name"]),
			"date":  date,
		})
	}
	return out
}

func trimCompare(data map[string]interface{}) interface{} {
	list, _ := data["list"].([]interface{})
	var out []interface{}
	for _, p := range list {
		pMap, _ := p.(map[string]interface{})
		out = append(out, pick(pMap, "name", "counter_id", "price_close", "market_value",
			"pe", "pb", "ps", "roe", "roa", "net_margin", "div_yld", "eps", "sales", "net_income"))
	}
	return out
}

func getFloat(m map[string]interface{}, key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	}
	return 0, false
}

func maxFloat(vals []float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func minFloat(vals []float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func firstNonNil(vals ...interface{}) interface{} {
	for _, v := range vals {
		if v != nil && v != "" {
			return v
		}
	}
	return nil
}

// ── Output ────────────────────────────────────────────────────────────

func section(outDir, title, name string, trimFn func(map[string]interface{}) interface{}) {
	fmt.Printf("===== %s =====\n", title)
	rawPath := filepath.Join(outDir, name+".json")
	errPath := filepath.Join(outDir, name+".err")

	data, err := os.ReadFile(rawPath)
	if err == nil {
		var parsed map[string]interface{}
		if json.Unmarshal(data, &parsed) == nil {
			var trimmed interface{}
			if trimFn != nil {
				trimmed = slim(trimFn(parsed))
			} else {
				trimmed = slim(parsed)
			}
			out, _ := json.Marshal(trimmed)
			fmt.Println(string(out))
			return
		}
		fmt.Printf("N/A (trim failed; raw: %s)\n", rawPath)
		return
	}

	errData, err2 := os.ReadFile(errPath)
	if err2 == nil {
		msg := string(errData)
		if len(msg) > 200 {
			msg = msg[:200]
		}
		msg = strings.ReplaceAll(msg, "\n", " ")
		fmt.Printf("N/A (%s)\n", msg)
		return
	}

	fmt.Println("N/A (no data)")
}

func main() {
	full := flag.Bool("full", false, "include balance sheet, cash flow, filings, industry valuation, peer compare")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		die("usage: longbridge-earnings-collect <SYMBOL> [--full]", 2)
	}

	symbol := strings.TrimSpace(args[0])

	// HK symbols: strip leading zeros
	if strings.HasSuffix(strings.ToUpper(symbol), ".HK") {
		symbol = strings.TrimLeft(symbol, "0")
	}

	if _, err := exec.LookPath("longbridge"); err != nil {
		die("ERROR: longbridge CLI not found. Install from https://open.longbridge.com", 3)
	}

	outDir := filepath.Join(os.TempDir(), fmt.Sprintf("lb_earnings_%s", strings.ReplaceAll(strings.ToLower(symbol), ".", "_")))
	os.MkdirAll(outDir, 0755)

	// Lite jobs (always)
	jobs := map[string][]string{
		"snapshot":     {"financial-report", "snapshot", symbol},
		"is_qf":        {"financial-report", symbol, "--kind", "IS", "--report", "qf"},
		"consensus":    {"consensus", symbol},
		"forecast_eps": {"forecast-eps", symbol},
		"quote":        {"quote", symbol},
		"calc_index":   {"calc-index", symbol},
		"rating":       {"institution-rating", symbol},
		"segments":     {"business-segments", symbol},
		"news":         {"news", symbol, "--count", "10"},
		"kline":        {"kline", symbol, "--period", "day", "--count", "250"},
	}

	if *full {
		jobs["bs_qf"] = []string{"financial-report", symbol, "--kind", "BS", "--report", "qf"}
		jobs["cf_qf"] = []string{"financial-report", symbol, "--kind", "CF", "--report", "qf"}
		jobs["filings"] = []string{"filing", symbol, "--count", "10"}
		jobs["ind_val"] = []string{"industry-valuation", "dist", symbol}
		jobs["compare"] = []string{"compare", symbol}
		jobs["rating_his"] = []string{"institution-rating", symbol, "--history"}
	}

	var wg sync.WaitGroup
	for name, cliArgs := range jobs {
		wg.Add(1)
		go func(n string, a []string) {
			defer wg.Done()
			fetch(outDir, n, a)
		}(name, cliArgs)
	}
	wg.Wait()

	// Quarterly statement retry for semi-annual reporters
	isFile := filepath.Join(outDir, "is_qf.json")
	data, err := os.ReadFile(isFile)
	empty := true
	if err == nil {
		var parsed map[string]interface{}
		if json.Unmarshal(data, &parsed) == nil {
			trimmed := trimStatement(parsed)
			if arr, ok := trimmed.([]interface{}); ok {
				empty = len(arr) == 0
			}
		}
	}
	if empty {
		fetch(outDir, "is_qf", []string{"financial-report", symbol, "--kind", "IS", "--report", "saf"})
	}

	fmt.Printf("SYMBOL: %s\n", symbol)
	fmt.Printf("COLLECTED_AT: %s\n", time.Now().Format("2006-01-02 15:04 MST"))
	fmt.Printf("RAW_DIR: %s  (full statements/filings live here — reuse, do not re-fetch)\n", outDir)

	section(outDir, "SNAPSHOT (latest period)", "snapshot", dropEmpty)
	section(outDir, "INCOME_STATEMENT (last 8 quarters)", "is_qf", trimStatement)
	section(outDir, "CONSENSUS (estimate vs actual, recent periods)", "consensus", trimConsensus)
	section(outDir, "FORECAST_EPS (annual consensus range, latest 3 windows)", "forecast_eps", trimForecastEPS)
	section(outDir, "QUOTE", "quote", nil)
	section(outDir, "CALC_INDEX (PE/PB/mktcap)", "calc_index", nil)
	section(outDir, "INSTITUTION_RATING", "rating", nil)
	section(outDir, "SEGMENTS (revenue breakdown)", "segments", nil)
	section(outDir, "NEWS (latest 10 headlines)", "news", trimNews)
	section(outDir, "KLINE (20 recent closes + 250d range)", "kline", trimKline)

	if *full {
		section(outDir, "FILINGS (latest 10)", "filings", trimFilings)
		section(outDir, "INDUSTRY_VALUATION (percentile dist)", "ind_val", nil)
		section(outDir, "PEER_COMPARE", "compare", trimCompare)
		fmt.Printf("===== FULL-MODE RAW FILES =====\n")
		fmt.Printf("BS/CF statements, rating history: %s/{bs_qf,cf_qf,rating_his}.json\n", outDir)
	}
}
