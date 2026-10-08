package pkg

import (
	"fmt"
	"os"
	"regexp"
)

func RenderSVG(path string, stats GitHubStats, age int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read SVG: %w", err)
	}

	svg := string(data)

	values := map[string]string{
		"age_data":      fmt.Sprintf("%d years", age),
		"repo_data":     fmt.Sprintf("%d", stats.Repositories),
		"contrib_data":  fmt.Sprintf("%d", stats.Contributions),
		"star_data":     fmt.Sprintf("%d", stats.Stars),
		"commit_data":   fmt.Sprintf("%d", stats.Commits),
		"follower_data": fmt.Sprintf("%d", stats.Followers),

		"loc_data": fmt.Sprintf("%d", stats.LinesOfCode),
		"loc_add":  fmt.Sprintf("%d", stats.LinesAdded),
		"loc_del":  fmt.Sprintf("%d", stats.LinesDeleted),
	}

	for id, value := range values {
		svg = replaceTspan(svg, id, value)
	}

	if err := os.WriteFile(path, []byte(svg), 0644); err != nil {
		return fmt.Errorf("write SVG: %w", err)
	}

	return nil
}

func replaceTspan(svg, id, value string) string {
	pattern := regexp.MustCompile(`(<tspan\b[^>]*\bid="` + regexp.QuoteMeta(id) + `"[^>]*>)(.*?)(</tspan>)`)

	return pattern.ReplaceAllString(svg, "${1}"+value+"${3}")
}