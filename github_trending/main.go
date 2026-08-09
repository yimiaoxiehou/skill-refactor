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

var periodConfig = map[string]struct {
	Days  int
	Label string
	Emoji string
}{
	"daily":   {1, "今日热门", "🔥"},
	"weekly":  {7, "本周热门", "📊"},
	"monthly": {30, "本月热门", "📈"},
}

var languages = []string{
	"", "python", "javascript", "typescript", "go", "rust",
	"java", "cpp", "c", "swift", "kotlin", "ruby", "php",
}

type Repo struct {
	FullName        string   `json:"full_name"`
	HTMLURL         string   `json:"html_url"`
	StargazersCount int      `json:"stargazers_count"`
	ForksCount      int      `json:"forks_count"`
	Language        string   `json:"language"`
	Description     string   `json:"description"`
	Topics          []string `json:"topics"`
}

type searchResult struct {
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
	req.Header.Set("User-Agent", "github-trending-cn-go/2.0")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result searchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func fetchTrending(period string, limit int, language string, token string) []Repo {
	days := 7
	if p, ok := periodConfig[period]; ok {
		days = p.Days
	}
	since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour).Format("2006-01-02")

	seen := make(map[string]bool)
	var results []Repo

	langFilter := ""
	if language != "" {
		langFilter = " language:" + language
	}

	// First batch: pushed + stars sorted
	query := fmt.Sprintf("pushed:>=%s stars:>=10%s", since, langFilter)
	items, err := ghSearch(query, token)
	if err == nil {
		for _, item := range items {
			if !seen[item.FullName] {
				seen[item.FullName] = true
				results = append(results, item)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "[WARN] GitHub API error: %v\n", err)
	}

	// Supplement with top languages if no language filter
	if language == "" && len(results) < limit*2 {
		for _, lang := range languages[1:6] {
			if len(results) >= limit*3 {
				break
			}
			q := fmt.Sprintf("pushed:>=%s stars:>=50 language:%s", since, lang)
			items, err := ghSearch(q, token)
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

func formatOutput(repos []Repo, period string, language string) string {
	p := periodConfig[period]
	now := time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04 MST")
	langTag := ""
	if language != "" {
		langTag = " · " + language
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s **GitHub Trending — %s%s**\n", p.Emoji, p.Label, langTag))
	sb.WriteString(fmt.Sprintf("数据时间：%s  |  共 %d 个项目\n\n", now, len(repos)))

	for i, r := range repos {
		stars := fmtNum(r.StargazersCount)
		forks := fmtNum(r.ForksCount)
		lang := r.Language
		if lang == "" {
			lang = "N/A"
		}
		desc := r.Description
		if len([]rune(desc)) > 100 {
			desc = string([]rune(desc)[:97]) + "..."
		}

		var topicStr string
		topics := r.Topics
		if len(topics) > 3 {
			topics = topics[:3]
		}
		if len(topics) > 0 {
			for i, t := range topics {
				topics[i] = "`" + t + "`"
			}
			topicStr = "  " + strings.Join(topics, " ")
		}

		sb.WriteString(fmt.Sprintf("**#%d** [%s](%s)\n", i+1, r.FullName, r.HTMLURL))
		sb.WriteString(fmt.Sprintf("⭐ %s  🍴 %s  🔤 %s%s\n", stars, forks, lang, topicStr))
		if desc != "" {
			sb.WriteString(fmt.Sprintf("> %s\n", desc))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func main() {
	period := flag.String("period", "daily", "daily|weekly|monthly")
	limit := flag.Int("limit", 20, "number of repos")
	language := flag.String("language", "", "programming language filter")
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub token")
	jsonOut := flag.Bool("json", false, "output raw JSON")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "[INFO] 正在获取 GitHub %s trending...\n", *period)

	repos := fetchTrending(*period, *limit, *language, *token)

	if len(repos) == 0 {
		fmt.Fprintln(os.Stderr, "[ERROR] 未获取到数据，请检查网络或 GitHub API 限额")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "[INFO] 获取到 %d 个项目\n", len(repos))

	if *jsonOut {
		type jsonRepo struct {
			Rank        int      `json:"rank"`
			Name        string   `json:"name"`
			URL         string   `json:"url"`
			Description string   `json:"description"`
			Stars       int      `json:"stars"`
			Forks       int      `json:"forks"`
			Language    string   `json:"language"`
			Topics      []string `json:"topics"`
		}
		output := make([]jsonRepo, len(repos))
		for i, r := range repos {
			output[i] = jsonRepo{
				Rank: i + 1, Name: r.FullName, URL: r.HTMLURL,
				Description: r.Description, Stars: r.StargazersCount,
				Forks: r.ForksCount, Language: r.Language, Topics: r.Topics,
			}
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Print(formatOutput(repos, *period, *language))
	}
}
