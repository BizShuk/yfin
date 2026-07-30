package unit

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
var markdownFencePattern = regexp.MustCompile("(?s)```(?:bash|sh)?\\s*(.*?)```")

func TestCurrentDocsDoNotReferenceRemovedCLISurfaces(t *testing.T) {
	root := repositoryRoot(t)
	banned := []string{
		"config-effective",
		"--preview-proto",
		"--publish",
		"yfin bus",
		"dashboards/runbooks",
		"config/types/",
		"svc/norm/security.go",
		"NewScrapeClientWithHTTP",
		"yfin_scrape_requests_total",
		"yfin_fallback_decisions_total",
	}

	for _, path := range currentMarkdownFiles(t, root) {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		for _, token := range banned {
			require.NotContainsf(t, string(content), token, "%s contains removed surface %q", path, token)
		}
		require.NotRegexpf(t, regexp.MustCompile(`\binf\b`), string(content), "%s claims an external inf observability backend owned by yfin", path)
		require.NotRegexpf(t, regexp.MustCompile(`(?m)yfin[^\n]*\bpull\b[^\n]*--preview\b`), string(content), "%s documents removed pull --preview flag", path)
		require.NotRegexpf(t, regexp.MustCompile(`(?m)yfin[^\n]*\bquote\b[^\n]*--preview\b`), string(content), "%s documents removed quote --preview flag", path)
		require.NotRegexpf(t, regexp.MustCompile(`(?m)go build [^\n]*\./cmd/yfin\b`), string(content), "%s documents the removed cmd/yfin entrypoint", path)
		require.NotRegexpf(t, regexp.MustCompile(`(?m)^\s*make (?:build|test|test-coverage|lint|fmt)\b`), string(content), "%s documents a missing Makefile target", path)
		for _, match := range markdownFencePattern.FindAllStringSubmatch(string(content), -1) {
			if strings.Contains(match[1], "yfin scrape") {
				require.NotContainsf(t, match[1], "--fallback", "%s documents removed yfin scrape --fallback flag", path)
			}
			normalizedCommands := strings.ReplaceAll(match[1], "\\\n", " ")
			for _, line := range strings.Split(normalizedCommands, "\n") {
				if strings.Contains(line, "yfin") && strings.Contains(line, "scrape") {
					require.NotContainsf(t, line, "--concurrency", "%s claims yfin scrape consumes --concurrency", path)
				}
			}
		}
	}
}

func TestCurrentDocsLocalLinksResolve(t *testing.T) {
	root := repositoryRoot(t)

	for _, path := range currentMarkdownFiles(t, root) {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		for _, match := range markdownLinkPattern.FindAllStringSubmatch(string(content), -1) {
			target := strings.TrimSpace(strings.Trim(match[1], "<>"))
			if target == "" ||
				strings.HasPrefix(target, "#") ||
				strings.HasPrefix(target, "http://") ||
				strings.HasPrefix(target, "https://") ||
				strings.HasPrefix(target, "mailto:") {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			target = strings.SplitN(target, "?", 2)[0]
			if target == "" {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(target)))
			_, err := os.Stat(resolved)
			require.NoErrorf(t, err, "%s links to missing local target %s", path, target)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func currentMarkdownFiles(t *testing.T, root string) []string {
	t.Helper()
	paths := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, "CLAUDE.md"),
	}
	for _, directory := range []string{"docs", "monitoring"} {
		base := filepath.Join(root, directory)
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				relative, relativeErr := filepath.Rel(root, path)
				if relativeErr != nil {
					return relativeErr
				}
				switch filepath.ToSlash(relative) {
				case "docs/history", "docs/memory", "docs/specs":
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) == ".md" && filepath.Base(path) != "CHANGELOG.md" {
				paths = append(paths, path)
			}
			return nil
		})
		require.NoError(t, err)
	}
	return paths
}
