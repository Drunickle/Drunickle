package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type githubUser struct {
	PublicRepos int `json:"public_repos"`
	Followers   int `json:"followers"`
}

type githubRepository struct {
	StargazersCount int  `json:"stargazers_count"`
	Fork            bool `json:"fork"`
	Archived        bool `json:"archived"`
}

type graphqlResponse struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				TotalCommitContributions int `json:"totalCommitContributions"`

				TotalRepositoriesWithContributedCommits int `json:"totalRepositoriesWithContributedCommits"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
}

func FetchGitHubStats(username string) (GitHubStats, error) {
	var stats GitHubStats

	user, err := githubGET[githubUser](
		"https://api.github.com/users/"+username,
	)
	if err != nil {
		return stats, err
	}

	stats.Repositories = user.PublicRepos
	stats.Followers = user.Followers

	repositories, err := githubGET[[]githubRepository](
		"https://api.github.com/users/"+username+"/repos?per_page=100&type=owner",
	)
	if err != nil {
		return stats, err
	}

	for _, repo := range repositories {
		if repo.Fork || repo.Archived {
			continue
		}

		stats.Stars += repo.StargazersCount
	}

	token := os.Getenv("GITHUB_TOKEN")

	if token != "" {
		contributions, err := fetchContributions(username, token)
		if err != nil {
			return stats, err
		}

		stats.Commits = contributions.Commits
		stats.Contributions = contributions.Contributions
	}

	return stats, nil
}

func fetchContributions(username, token string) (GitHubStats, error) {
	var stats GitHubStats

	query := `
	query($login: String!) {
		user(login: $login) {
			contributionsCollection {
				totalCommitContributions
				totalRepositoriesWithContributedCommits
			}
		}
	}
	`

	body := map[string]interface{}{
		"query": query,
		"variables": map[string]string{
			"login": username,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return stats, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.github.com/graphql",
		bytes.NewReader(payload),
	)
	if err != nil {
		return stats, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stats, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)

		return stats, fmt.Errorf(
			"GitHub GraphQL returned %s: %s",
			resp.Status,
			string(data),
		)
	}

	var result graphqlResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return stats, err
	}

	stats.Commits =
		result.Data.User.ContributionsCollection.TotalCommitContributions

	stats.Contributions =
		result.Data.User.ContributionsCollection.TotalRepositoriesWithContributedCommits

	return stats, nil
}

func githubGET[T any](url string) (T, error) {
	var result T

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)

		return result, fmt.Errorf(
			"GitHub API returned %s: %s",
			resp.Status,
			string(data),
		)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}