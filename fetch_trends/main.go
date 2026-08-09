package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var periods = map[string]struct {
	Days  int
	Label string
	Emoji string
}{
	"daily":   {1, "日榜", "📅"},
	"weekly":  {7, "周榜", "📊"},
	"monthly": {30, "月榜", "📈"},
}

type Repo struct {
	FullName        string `json:"full_name"`
	HTMLURL         string `json:"html_url"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	Language        string `json:"language"`
	Description     string `json:"description"`
}

type SearchResult struct {
	Items []Repo `json:"items"`
}

func ghSearch(query string, token string) ([]Repo, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "stars")
	params.Set("order", "desc")
	params.Set("per_page", "30")

	req, _ := http.NewRequest("GET", "https://api.github.com/search/repositories?"+params.Encode(), nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "github-ai-trends-go")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func fetchTrending(period string, limit int, token string) []Repo {
	days := 7
	if p, ok := periods[period]; ok {
		days = p.Days
	}
	since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour).Format("2006-01-02")

	seen := make(map[string]bool)
	var results []Repo

	keywords := []string{"ai", "llm", "gpt", "agent", "transformer", "diffusion", "rag", "ml"}
	for _, kw := range keywords {
		if len(results) >= limit*2 {
			break
		}
		items, err := ghSearch(fmt.Sprintf("%s in:name,description pushed:>=%s stars:>=10", kw, since), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] GitHub API error: %v\n", err)
			continue
		}
		for _, item := range items {
			if !seen[item.FullName] {
				seen[item.FullName] = true
				results = append(results, item)
			}
		}
	}

	topics := []string{"artificial-intelligence", "llm", "generative-ai", "ai-agent"}
	for _, topic := range topics {
		if len(results) >= limit*3 {
			break
		}
		items, err := ghSearch(fmt.Sprintf("topic:%s pushed:>=%s stars:>=10", topic, since), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] GitHub API error: %v\n", err)
			continue
		}
		for _, item := range items {
			if !seen[item.FullName] {
				seen[item.FullName] = true
				results = append(results, item)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StargazersCount > results[j].StargazersCount
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func fmtNum(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func formatOutput(repos []Repo, period string) string {
	p := periods[period]
	now := time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s **GitHub AI 趋势榜 — %s**\n", p.Emoji, p.Label))
	sb.WriteString(fmt.Sprintf("生成时间：%s\n\n", now))

	for i, r := range repos {
		stars := fmtNum(r.StargazersCount)
		forks := fmtNum(r.ForksCount)
		lang := r.Language
		if lang == "" {
			lang = "N/A"
		}
		desc := r.Description
		if len([]rune(desc)) > 80 {
			desc = string([]rune(desc)[:77]) + "..."
		}

		sb.WriteString(fmt.Sprintf("**#%d** [%s](%s)\n", i+1, r.FullName, r.HTMLURL))
		sb.WriteString(fmt.Sprintf("⭐ %s · 🍴 %s · %s\n", stars, forks, lang))
		if desc != "" {
			sb.WriteString(fmt.Sprintf("_%s_\n", desc))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func main() {
	period := flag.String("period", "weekly", "daily|weekly|monthly")
	limit := flag.Int("limit", 20, "number of repos")
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub token")
	jsonOut := flag.Bool("json", false, "output raw JSON")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "Fetching %s AI trends...\n", *period)

	if _, ok := periods[*period]; !ok {
		fmt.Fprintf(os.Stderr, "Invalid period: %s\n", *period)
		os.Exit(1)
	}

	repos := fetchTrending(*period, *limit, *token)

	if len(repos) == 0 {
		fmt.Fprintln(os.Stderr, "No repos found.")
		os.Exit(1)
	}

	if *jsonOut {
		type jsonRepo struct {
			Rank        int    `json:"rank"`
			Name        string `json:"name"`
			URL         string `json:"url"`
			Stars       int    `json:"stars"`
			Forks       int    `json:"forks"`
			Language    string `json:"language"`
			Description string `json:"description"`
		}
		output := make([]jsonRepo, len(repos))
		for i, r := range repos {
			output[i] = jsonRepo{
				Rank: i + 1, Name: r.FullName, URL: r.HTMLURL,
				Stars: r.StargazersCount, Forks: r.ForksCount,
				Language: r.Language, Description: r.Description,
			}
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Print(formatOutput(repos, *period))
	}
}
