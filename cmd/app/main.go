package main

import (
	"fmt"
	"os"
	"time"

	"drunickle/drunickle/pkg"
)

const (
	svgPath = "assets/theme.svg"

	birthYear  = 2005
	birthMonth = time.May
	birthDay   = 5
)

func main() {
	username := os.Getenv("GITHUB_USERNAME")

	if username == "" {
		fmt.Println("GITHUB_USERNAME is not set")
		os.Exit(1)
	}

	fmt.Println("Fetching GitHub statistics...")

	stats, err := pkg.FetchGitHubStats(username)
	if err != nil {
		fmt.Println("GitHub error:", err)
		os.Exit(1)
	}

	age := calculateAge(
		birthYear,
		birthMonth,
		birthDay,
		time.Now(),
	)

	if err := pkg.RenderSVG(
		svgPath,
		stats,
		age,
	); err != nil {
		fmt.Println("Renderer error:", err)
		os.Exit(1)
	}

	fmt.Println("Profile updated successfully.")
}

func calculateAge(
	year int,
	month time.Month,
	day int,
	now time.Time,
) int {
	birthDate := time.Date(
		year,
		month,
		day,
		0,
		0,
		0,
		0,
		now.Location(),
	)

	age := now.Year() - birthDate.Year()

	if now.Before(birthDate.AddDate(age, 0, 0)) {
		age--
	}

	return age
}
