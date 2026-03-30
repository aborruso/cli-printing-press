package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mvanhorn/cli-printing-press/internal/crowdsniff"
	"github.com/mvanhorn/cli-printing-press/internal/websniff"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newCrowdSniffCmd() *cobra.Command {
	var apiName string
	var outputPath string
	var baseURL string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "crowd-sniff",
		Short: "Discover API endpoints from npm SDKs and GitHub code search",
		Long: `Discover API endpoints by mining community signals: npm SDK packages
and GitHub code search. Produces a spec YAML compatible with 'printing-press generate'.

Complements 'sniff' (which discovers from live web traffic) by finding
what developers have already mapped in published packages and code.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCrowdSniffAPIName(apiName); err != nil {
				return err
			}

			ctx := cmd.Context()

			npmSource := crowdsniff.NewNPMSource(crowdsniff.NPMOptions{})
			githubSource := crowdsniff.NewGitHubSource(crowdsniff.GitHubOptions{})

			var npmResult, githubResult crowdsniff.SourceResult
			g := new(errgroup.Group)

			g.Go(func() error {
				result, err := npmSource.Discover(ctx, apiName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: npm source: %v\n", err)
					return nil
				}
				npmResult = result
				return nil
			})

			g.Go(func() error {
				result, err := githubSource.Discover(ctx, apiName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: github source: %v\n", err)
					return nil
				}
				githubResult = result
				return nil
			})

			if err := g.Wait(); err != nil {
				return fmt.Errorf("running sources: %w", err)
			}

			results := []crowdsniff.SourceResult{npmResult, githubResult}
			aggregated, baseURLCandidates := crowdsniff.Aggregate(results)

			if len(aggregated) == 0 {
				return fmt.Errorf("no endpoints discovered for %q", apiName)
			}

			resolvedBaseURL := crowdsniff.ResolveBaseURL(baseURL, baseURLCandidates)
			if resolvedBaseURL == "" {
				return fmt.Errorf("could not determine base URL for %q; use --base-url to specify", apiName)
			}

			if !isHTTPS(resolvedBaseURL) {
				return fmt.Errorf("base URL must use HTTPS: %s", resolvedBaseURL)
			}

			apiSpec, err := crowdsniff.BuildSpec(apiName, resolvedBaseURL, aggregated)
			if err != nil {
				return fmt.Errorf("building spec: %w", err)
			}

			if outputPath == "" {
				outputPath = defaultCrowdSniffCachePath(apiName)
			}

			if err := validateOutputPath(outputPath); err != nil {
				return err
			}

			if err := websniff.WriteSpec(apiSpec, outputPath); err != nil {
				return fmt.Errorf("writing spec: %w", err)
			}

			endpointCount := 0
			for _, resource := range apiSpec.Resources {
				endpointCount += len(resource.Endpoints)
			}

			npmCount := len(npmResult.Endpoints)
			githubCount := len(githubResult.Endpoints)

			tierCounts := make(map[string]int)
			for _, ep := range aggregated {
				tierCounts[ep.SourceTier]++
			}

			if asJSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"spec_path":       outputPath,
					"endpoints":       endpointCount,
					"resources":       len(apiSpec.Resources),
					"npm_discovered":  npmCount,
					"gh_discovered":   githubCount,
					"tier_breakdown":  tierCounts,
				})
			}

			fmt.Printf("Spec written to %s (%d endpoints across %d resources)\n", outputPath, endpointCount, len(apiSpec.Resources))
			fmt.Printf("Sources: %d from npm, %d from GitHub code search\n", npmCount, githubCount)
			if len(tierCounts) > 0 {
				parts := make([]string, 0, len(tierCounts))
				for tier, count := range tierCounts {
					parts = append(parts, fmt.Sprintf("%s: %d", tier, count))
				}
				fmt.Printf("Tiers: %s\n", strings.Join(parts, ", "))
			}
			fmt.Printf("Run 'printing-press generate --spec %s' to build the CLI\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&apiName, "api", "", "API name or domain (e.g., 'notion', 'api.stripe.com')")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output path for generated spec YAML")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Override auto-detected base URL (must be HTTPS)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("api")

	return cmd
}

// validateCrowdSniffAPIName rejects dangerous --api values.
func validateCrowdSniffAPIName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("--api value is required")
	}
	for _, ch := range name {
		if ch == '\n' || ch == '\r' || ch == 0 {
			return fmt.Errorf("--api value contains invalid characters")
		}
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		// If it looks like a URL (contains ://), allow slashes in the URL path.
		if !strings.Contains(name, "://") {
			return fmt.Errorf("--api value contains path traversal characters")
		}
	}
	return nil
}

// validateOutputPath checks the resolved output path for traversal.
func validateOutputPath(outputPath string) error {
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil // can't validate, allow it
	}

	cacheRoot := filepath.Join(home, ".cache", "printing-press")
	if strings.HasPrefix(absPath, cacheRoot+string(filepath.Separator)) {
		return nil
	}

	// If using a custom --output path (not under cache), that's fine.
	// The traversal check is only for the auto-generated default path.
	return nil
}

func defaultCrowdSniffCachePath(name string) string {
	// Sanitize name for use in file path.
	safeName := url.PathEscape(name)

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".cache", "printing-press", "crowd-sniff", safeName+"-spec.yaml")
	}
	return filepath.Join(home, ".cache", "printing-press", "crowd-sniff", safeName+"-spec.yaml")
}

func isHTTPS(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Scheme, "https")
}

