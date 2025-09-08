// Package cli/history provides intelligent git history navigation and search
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gajeshbhat/git-assist/internal/ai"
	"github.com/gajeshbhat/git-assist/internal/git"
	"github.com/spf13/cobra"
)

// Command flags for history
var (
	historySearch   string
	historyExplain  bool
	historyTimeline bool
	historyFeature  string
	historyAuthor   string
	historyCount    int
	historyFormat   string
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Intelligent git history navigation and search",
	Long: `Navigate and search git history with AI-powered insights.
Find commits, understand development patterns, and explore the story
behind your codebase.

Features:
• Smart commit search by content, author, or message
• Explain development story and patterns
• Timeline view of feature development
• Find commits related to specific functionality
• Analyze commit patterns and trends`,
	Example: `  git-assist history                    # Show recent history with insights
  git-assist history --search "auth"    # Search commits containing "auth"
  git-assist history --author john      # Show commits by author
  git-assist history --explain          # Explain development story
  git-assist history --timeline         # Show development timeline
  git-assist history --feature auth     # Show feature development history`,
	RunE: runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)

	// Add flags
	historyCmd.Flags().StringVar(&historySearch, "search", "", "search commits by message, content, or metadata")
	historyCmd.Flags().BoolVar(&historyExplain, "explain", false, "explain development story and patterns")
	historyCmd.Flags().BoolVar(&historyTimeline, "timeline", false, "show development timeline")
	historyCmd.Flags().StringVar(&historyFeature, "feature", "", "show development history for specific feature")
	historyCmd.Flags().StringVar(&historyAuthor, "author", "", "filter commits by author")
	historyCmd.Flags().IntVar(&historyCount, "count", 20, "number of commits to show")
	historyCmd.Flags().StringVar(&historyFormat, "format", "detailed", "output format (detailed, compact, timeline)")
}

// runHistory executes the history command
func runHistory(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Verify we're in a git repository
	gitRepo := git.NewRepository(cwd)
	if !gitRepo.IsGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	// Handle different history operations
	if historySearch != "" {
		return searchCommitHistory(gitRepo, historySearch)
	}

	if historyExplain {
		return explainDevelopmentStory(gitRepo)
	}

	if historyTimeline {
		return showDevelopmentTimeline(gitRepo)
	}

	if historyFeature != "" {
		return showFeatureHistory(gitRepo, historyFeature)
	}

	// Default: show enhanced history
	return showEnhancedHistory(gitRepo)
}

// showEnhancedHistory shows git history with AI insights
func showEnhancedHistory(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("📚 Enhanced Git History")
		fmt.Println()
	}

	// Get recent commits
	commits, err := getDetailedCommits(gitRepo, historyCount, historyAuthor)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	if len(commits) == 0 {
		color.New(color.FgYellow).Println("⚠️  No commits found")
		return nil
	}

	// Show commits based on format
	switch historyFormat {
	case "compact":
		showCompactHistory(commits)
	case "timeline":
		showTimelineHistory(commits)
	default:
		showDetailedHistory(commits)
	}

	// Provide insights
	fmt.Println()
	showHistoryInsights(commits)

	return nil
}

// searchCommitHistory searches commits by various criteria
func searchCommitHistory(gitRepo *git.Repository, query string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("🔍 Searching for: '%s'\n", query)
		fmt.Println()
	}

	// Search in commit messages
	messageResults, err := searchCommitMessages(gitRepo, query)
	if err != nil {
		return fmt.Errorf("failed to search commit messages: %w", err)
	}

	// Search in commit content
	contentResults, err := searchCommitContent(gitRepo, query)
	if err != nil {
		return fmt.Errorf("failed to search commit content: %w", err)
	}

	// Combine and deduplicate results
	allResults := combineSearchResults(messageResults, contentResults)

	if len(allResults) == 0 {
		color.New(color.FgYellow).Printf("⚠️  No commits found matching '%s'\n", query)
		return nil
	}

	color.New(color.FgGreen, color.Bold).Printf("📋 Found %d matching commits:\n", len(allResults))
	fmt.Println()

	for i, commit := range allResults {
		if i >= 10 { // Limit results
			color.New(color.FgBlue).Printf("   ... and %d more (use --count to see more)\n", len(allResults)-i)
			break
		}

		showSearchResult(commit, query)
	}

	return nil
}

// explainDevelopmentStory uses AI to explain the development story
func explainDevelopmentStory(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("📖 Development Story")
		fmt.Println()
	}

	// Get recent commits for analysis
	commits, err := getDetailedCommits(gitRepo, 50, "")
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		color.New(color.FgYellow).Println("⚠️  No commits found")
		return nil
	}

	// Try AI explanation first
	if err := explainStoryWithAI(commits); err != nil {
		// Fallback to rule-based analysis
		explainStoryRuleBased(commits)
	}

	return nil
}

// showDevelopmentTimeline shows a timeline view of development
func showDevelopmentTimeline(gitRepo *git.Repository) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Println("📅 Development Timeline")
		fmt.Println()
	}

	// Get commits grouped by time periods
	commits, err := getDetailedCommits(gitRepo, historyCount, historyAuthor)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		color.New(color.FgYellow).Println("⚠️  No commits found")
		return nil
	}

	// Group commits by time periods
	timeGroups := groupCommitsByTime(commits)

	for period, periodCommits := range timeGroups {
		color.New(color.FgBlue, color.Bold).Printf("📅 %s (%d commits)\n", period, len(periodCommits))

		for _, commit := range periodCommits {
			fmt.Printf("   %s %s - %s\n",
				commit.Date.Format("15:04"),
				commit.Hash[:8],
				truncateString(commit.Message, 60))
		}
		fmt.Println()
	}

	return nil
}

// showFeatureHistory shows development history for a specific feature
func showFeatureHistory(gitRepo *git.Repository, feature string) error {
	if !quiet {
		color.New(color.FgCyan, color.Bold).Printf("🎯 Feature History: %s\n", feature)
		fmt.Println()
	}

	// Search for commits related to the feature
	commits, err := searchCommitMessages(gitRepo, feature)
	if err != nil {
		return fmt.Errorf("failed to search feature commits: %w", err)
	}

	if len(commits) == 0 {
		color.New(color.FgYellow).Printf("⚠️  No commits found for feature '%s'\n", feature)
		return nil
	}

	color.New(color.FgGreen, color.Bold).Printf("📋 Found %d commits for '%s':\n", len(commits), feature)
	fmt.Println()

	// Show feature development timeline
	for i, commit := range commits {
		showFeatureCommit(commit, i+1, len(commits))
	}

	// Analyze feature development pattern
	fmt.Println()
	analyzeFeatureDevelopment(commits, feature)

	return nil
}

// DetailedCommit represents a commit with full information
type DetailedCommit struct {
	Hash       string
	Message    string
	Author     string
	Date       time.Time
	Files      []string
	Insertions int
	Deletions  int
}

// getDetailedCommits gets commits with detailed information
func getDetailedCommits(gitRepo *git.Repository, count int, author string) ([]DetailedCommit, error) {
	args := []string{"log", "--pretty=format:%H|%s|%an|%ad", "--date=iso", "--numstat"}

	if count > 0 {
		args = append(args, fmt.Sprintf("-%d", count))
	}

	if author != "" {
		args = append(args, fmt.Sprintf("--author=%s", author))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseDetailedCommits(string(output)), nil
}

// parseDetailedCommits parses git log output with numstat
func parseDetailedCommits(output string) []DetailedCommit {
	var commits []DetailedCommit
	lines := strings.Split(output, "\n")

	var currentCommit *DetailedCommit

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a commit line (contains |)
		if strings.Contains(line, "|") && len(strings.Split(line, "|")) == 4 {
			// Save previous commit if exists
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}

			// Parse new commit
			parts := strings.Split(line, "|")
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])

			currentCommit = &DetailedCommit{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
				Date:    date,
				Files:   []string{},
			}
		} else if currentCommit != nil {
			// This is a numstat line (insertions, deletions, filename)
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				filename := parts[2]
				currentCommit.Files = append(currentCommit.Files, filename)

				// Parse insertions/deletions if they're numbers
				if parts[0] != "-" {
					if insertions := parseInt(parts[0]); insertions > 0 {
						currentCommit.Insertions += insertions
					}
				}
				if parts[1] != "-" {
					if deletions := parseInt(parts[1]); deletions > 0 {
						currentCommit.Deletions += deletions
					}
				}
			}
		}
	}

	// Add the last commit
	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}

	return commits
}

// showDetailedHistory shows commits in detailed format
func showDetailedHistory(commits []DetailedCommit) {
	for i, commit := range commits {
		// Commit header
		color.New(color.FgYellow, color.Bold).Printf("📝 %s", commit.Hash[:8])
		color.New(color.FgGreen).Printf(" by %s", commit.Author)
		color.New(color.FgBlue).Printf(" %s\n", formatTimeAgo(commit.Date))

		// Commit message
		fmt.Printf("   %s\n", commit.Message)

		// File changes
		if len(commit.Files) > 0 {
			color.New(color.FgCyan).Printf("   📁 %d files changed", len(commit.Files))
			if commit.Insertions > 0 || commit.Deletions > 0 {
				color.New(color.FgGreen).Printf(" (+%d", commit.Insertions)
				color.New(color.FgRed).Printf(" -%d)", commit.Deletions)
			}
			fmt.Println()

			// Show first few files
			for j, file := range commit.Files {
				if j >= 3 {
					fmt.Printf("   ... and %d more files\n", len(commit.Files)-j)
					break
				}
				fmt.Printf("   • %s\n", file)
			}
		}

		// Separator
		if i < len(commits)-1 {
			fmt.Println()
		}
	}
}

// showCompactHistory shows commits in compact format
func showCompactHistory(commits []DetailedCommit) {
	for _, commit := range commits {
		color.New(color.FgYellow).Printf("%s ", commit.Hash[:8])
		color.New(color.FgBlue).Printf("%s ", formatTimeAgo(commit.Date))
		fmt.Printf("%s\n", truncateString(commit.Message, 60))
	}
}

// showTimelineHistory shows commits grouped by time
func showTimelineHistory(commits []DetailedCommit) {
	timeGroups := groupCommitsByTime(commits)

	for period, periodCommits := range timeGroups {
		color.New(color.FgBlue, color.Bold).Printf("📅 %s\n", period)

		for _, commit := range periodCommits {
			color.New(color.FgYellow).Printf("   %s ", commit.Hash[:8])
			fmt.Printf("%s\n", truncateString(commit.Message, 50))
		}
		fmt.Println()
	}
}

// showHistoryInsights provides insights about the commit history
func showHistoryInsights(commits []DetailedCommit) {
	if len(commits) == 0 {
		return
	}

	color.New(color.FgMagenta, color.Bold).Println("💡 Repository Insights:")

	// Activity analysis
	totalFiles := 0
	totalInsertions := 0
	totalDeletions := 0
	authors := make(map[string]int)

	for _, commit := range commits {
		totalFiles += len(commit.Files)
		totalInsertions += commit.Insertions
		totalDeletions += commit.Deletions
		authors[commit.Author]++
	}

	fmt.Printf("   📊 Activity: %d commits, %d files changed\n", len(commits), totalFiles)
	fmt.Printf("   📈 Changes: +%d -%d lines\n", totalInsertions, totalDeletions)
	fmt.Printf("   👥 Contributors: %d unique authors\n", len(authors))

	// Time analysis
	if len(commits) > 1 {
		timeSpan := commits[0].Date.Sub(commits[len(commits)-1].Date)
		fmt.Printf("   ⏰ Time span: %s\n", formatDuration(timeSpan))
	}

	// Most active author
	var mostActiveAuthor string
	maxCommits := 0
	for author, count := range authors {
		if count > maxCommits {
			maxCommits = count
			mostActiveAuthor = author
		}
	}

	if mostActiveAuthor != "" {
		fmt.Printf("   🏆 Most active: %s (%d commits)\n", mostActiveAuthor, maxCommits)
	}
}

// searchCommitMessages searches commit messages
func searchCommitMessages(gitRepo *git.Repository, query string) ([]DetailedCommit, error) {
	cmd := exec.Command("git", "log", "--grep="+query, "--pretty=format:%H|%s|%an|%ad", "--date=iso")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseSimpleCommits(string(output)), nil
}

// searchCommitContent searches commit content
func searchCommitContent(gitRepo *git.Repository, query string) ([]DetailedCommit, error) {
	cmd := exec.Command("git", "log", "-S"+query, "--pretty=format:%H|%s|%an|%ad", "--date=iso")
	cmd.Dir = gitRepo.Path()

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseSimpleCommits(string(output)), nil
}

// parseSimpleCommits parses simple git log output
func parseSimpleCommits(output string) []DetailedCommit {
	var commits []DetailedCommit
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])

			commits = append(commits, DetailedCommit{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
				Date:    date,
			})
		}
	}

	return commits
}

// combineSearchResults combines and deduplicates search results
func combineSearchResults(messageResults, contentResults []DetailedCommit) []DetailedCommit {
	seen := make(map[string]bool)
	var combined []DetailedCommit

	// Add message results
	for _, commit := range messageResults {
		if !seen[commit.Hash] {
			combined = append(combined, commit)
			seen[commit.Hash] = true
		}
	}

	// Add content results
	for _, commit := range contentResults {
		if !seen[commit.Hash] {
			combined = append(combined, commit)
			seen[commit.Hash] = true
		}
	}

	return combined
}

// showSearchResult shows a search result with highlighting
func showSearchResult(commit DetailedCommit, query string) {
	color.New(color.FgYellow).Printf("📝 %s ", commit.Hash[:8])
	color.New(color.FgGreen).Printf("by %s ", commit.Author)
	color.New(color.FgBlue).Printf("%s\n", formatTimeAgo(commit.Date))

	// Highlight query in message
	message := highlightQuery(commit.Message, query)
	fmt.Printf("   %s\n", message)
	fmt.Println()
}

// groupCommitsByTime groups commits by time periods
func groupCommitsByTime(commits []DetailedCommit) map[string][]DetailedCommit {
	groups := make(map[string][]DetailedCommit)

	for _, commit := range commits {
		period := getTimePeriod(commit.Date)
		groups[period] = append(groups[period], commit)
	}

	return groups
}

// getTimePeriod returns a time period string for a date
func getTimePeriod(date time.Time) string {
	now := time.Now()

	if date.After(now.AddDate(0, 0, -1)) {
		return "Today"
	} else if date.After(now.AddDate(0, 0, -7)) {
		return "This Week"
	} else if date.After(now.AddDate(0, -1, 0)) {
		return "This Month"
	} else if date.After(now.AddDate(-1, 0, 0)) {
		return "This Year"
	} else {
		return date.Format("2006")
	}
}

// showFeatureCommit shows a commit in feature development context
func showFeatureCommit(commit DetailedCommit, index, total int) {
	// Progress indicator
	color.New(color.FgBlue).Printf("[%d/%d] ", index, total)

	// Commit info
	color.New(color.FgYellow).Printf("%s ", commit.Hash[:8])
	color.New(color.FgGreen).Printf("%s ", formatTimeAgo(commit.Date))
	fmt.Printf("%s\n", commit.Message)
}

// analyzeFeatureDevelopment analyzes feature development patterns
func analyzeFeatureDevelopment(commits []DetailedCommit, feature string) {
	color.New(color.FgMagenta, color.Bold).Printf("📊 Feature Analysis: %s\n", feature)

	if len(commits) == 0 {
		return
	}

	// Time analysis
	if len(commits) > 1 {
		timeSpan := commits[0].Date.Sub(commits[len(commits)-1].Date)
		fmt.Printf("   ⏰ Development time: %s\n", formatDuration(timeSpan))
	}

	// Author analysis
	authors := make(map[string]int)
	for _, commit := range commits {
		authors[commit.Author]++
	}

	fmt.Printf("   👥 Contributors: %d\n", len(authors))
	fmt.Printf("   📝 Total commits: %d\n", len(commits))
}

// explainStoryWithAI uses AI to explain development story
func explainStoryWithAI(commits []DetailedCommit) error {
	// Create AI manager
	aiManager := ai.NewManager()

	// Try to setup AI
	err := aiManager.SetupFromConfig()
	if err != nil {
		return err
	}

	// Create story prompt
	commitSummary := ""
	for i, commit := range commits {
		if i >= 10 { // Limit to recent commits
			break
		}
		commitSummary += fmt.Sprintf("- %s: %s (by %s)\n",
			commit.Date.Format("Jan 2"), commit.Message, commit.Author)
	}

	prompt := fmt.Sprintf(`Analyze this git commit history and tell the development story:

Recent commits:
%s

Please explain:
1. What the main development themes are
2. What features or areas are being worked on
3. The development patterns you observe
4. Any insights about the project's direction

Keep it concise and insightful.`, commitSummary)

	// Generate story
	story, err := aiManager.GenerateCommitMessage(prompt)
	if err != nil {
		return err
	}

	// Display AI story
	color.New(color.FgGreen).Println("🤖 AI Development Story:")
	fmt.Println(story)

	return nil
}

// explainStoryRuleBased provides rule-based development story
func explainStoryRuleBased(commits []DetailedCommit) {
	color.New(color.FgGreen).Println("📚 Development Story (Rule-based):")

	if len(commits) == 0 {
		fmt.Println("   No commits to analyze")
		return
	}

	// Analyze commit patterns
	themes := make(map[string]int)
	authors := make(map[string]int)

	for _, commit := range commits {
		// Extract themes from commit messages
		message := strings.ToLower(commit.Message)
		if strings.Contains(message, "feat") || strings.Contains(message, "feature") {
			themes["features"]++
		} else if strings.Contains(message, "fix") || strings.Contains(message, "bug") {
			themes["fixes"]++
		} else if strings.Contains(message, "doc") {
			themes["documentation"]++
		} else if strings.Contains(message, "test") {
			themes["testing"]++
		} else if strings.Contains(message, "refactor") {
			themes["refactoring"]++
		} else {
			themes["other"]++
		}

		authors[commit.Author]++
	}

	// Report findings
	fmt.Printf("   📊 Analyzed %d commits\n", len(commits))

	fmt.Println("   🎯 Main themes:")
	for theme, count := range themes {
		if count > 0 {
			fmt.Printf("     • %s: %d commits\n", theme, count)
		}
	}

	fmt.Printf("   👥 Active contributors: %d\n", len(authors))

	if len(commits) > 1 {
		timeSpan := commits[0].Date.Sub(commits[len(commits)-1].Date)
		fmt.Printf("   ⏰ Recent activity span: %s\n", formatDuration(timeSpan))
	}
}

// Helper functions

// formatTimeAgo formats time as "X ago"
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Hour {
		return fmt.Sprintf("%.0fm ago", duration.Minutes())
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%.0fh ago", duration.Hours())
	} else if duration < 30*24*time.Hour {
		return fmt.Sprintf("%.0fd ago", duration.Hours()/24)
	} else {
		return t.Format("Jan 2, 2006")
	}
}

// formatDuration formats a duration in human-readable form
func formatDuration(d time.Duration) string {
	if d < 24*time.Hour {
		return fmt.Sprintf("%.0f hours", d.Hours())
	} else if d < 30*24*time.Hour {
		return fmt.Sprintf("%.0f days", d.Hours()/24)
	} else if d < 365*24*time.Hour {
		return fmt.Sprintf("%.0f months", d.Hours()/(24*30))
	} else {
		return fmt.Sprintf("%.1f years", d.Hours()/(24*365))
	}
}

// truncateString truncates a string to specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// highlightQuery highlights query terms in text (simple implementation)
func highlightQuery(text, query string) string {
	// Simple case-insensitive highlighting
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	if strings.Contains(lowerText, lowerQuery) {
		// For now, just return the text (highlighting would require more complex formatting)
		return text
	}

	return text
}

// parseInt safely parses an integer
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
