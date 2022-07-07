package github

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/github"
)

const (
	GITHUB_REPO  = "cli"
	GITHUB_OWNER = "groundcover-com"
)

type GitHubLatestRelease struct {
	assetId  int64
	assetUrl string
	Version  semver.Version
}

func NewGitHubLatestRelease(ctx context.Context) (release *GitHubLatestRelease, err error) {
	var githubRelease *github.RepositoryRelease

	client := github.NewClient(nil)
	if githubRelease, _, err = client.Repositories.GetLatestRelease(ctx, GITHUB_OWNER, GITHUB_REPO); err != nil {
		return
	}

	release = new(GitHubLatestRelease)
	if err = release.fetchVersion(githubRelease); err != nil {
		return
	}
	if err = release.fetchAsset(ctx, client, githubRelease); err != nil {
		return
	}
	return
}

func (release *GitHubLatestRelease) fetchVersion(githubRelease *github.RepositoryRelease) (err error) {
	release.Version, err = semver.ParseTolerant(githubRelease.GetTagName())
	return
}

func (release *GitHubLatestRelease) fetchAsset(ctx context.Context, client *github.Client, githubRelease *github.RepositoryRelease) (err error) {
	assetSuffix := fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, asset := range githubRelease.Assets {
		if strings.HasSuffix(asset.GetName(), assetSuffix) {
			release.assetId = asset.GetID()
			_, release.assetUrl, err = client.Repositories.DownloadReleaseAsset(ctx, GITHUB_OWNER, GITHUB_REPO, release.assetId)
			return
		}
	}
	return fmt.Errorf("failed to find asset for %s", assetSuffix)
}

func (release *GitHubLatestRelease) DownloadAsset(ctx context.Context) (*http.Response, error) {
	return http.Get(release.assetUrl)
}
