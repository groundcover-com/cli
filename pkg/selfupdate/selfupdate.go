package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/github"
	"github.com/minio/selfupdate"
)

type SelfUpdater struct {
	assetId     int64
	githubOwner string
	githubRepo  string
	assetUrl    string
	Version     semver.Version
}

func NewSelfUpdater(ctx context.Context, githubOwner, githubRepo string) (*SelfUpdater, error) {
	var err error
	var githubRelease *github.RepositoryRelease

	selfUpdater := new(SelfUpdater)
	selfUpdater.githubOwner = githubOwner
	selfUpdater.githubRepo = githubRepo

	client := github.NewClient(nil)
	if githubRelease, _, err = client.Repositories.GetLatestRelease(ctx, githubOwner, githubRepo); err != nil {
		return nil, err
	}
	if err = selfUpdater.fetchVersion(githubRelease); err != nil {
		return nil, err
	}
	if err = selfUpdater.fetchAsset(ctx, client, githubRelease); err != nil {
		return nil, err
	}
	return selfUpdater, err
}

func (selfUpdater *SelfUpdater) fetchVersion(githubRelease *github.RepositoryRelease) error {
	var err error

	selfUpdater.Version, err = semver.ParseTolerant(githubRelease.GetTagName())
	return err
}

func (selfUpdater *SelfUpdater) fetchAsset(ctx context.Context, client *github.Client, githubRelease *github.RepositoryRelease) error {
	var err error

	assetSuffix := fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, asset := range githubRelease.Assets {
		if strings.HasSuffix(asset.GetName(), assetSuffix) {
			selfUpdater.assetId = asset.GetID()
			_, selfUpdater.assetUrl, err = client.Repositories.DownloadReleaseAsset(ctx, selfUpdater.githubOwner, selfUpdater.githubRepo, selfUpdater.assetId)
			return err
		}
	}
	return fmt.Errorf("failed to find asset for %s", assetSuffix)
}

func (selfUpdater *SelfUpdater) IsLatestNewer(currentVersion semver.Version) bool {
	return selfUpdater.Version.Compare(currentVersion) > 0
}

func (selfUpdater *SelfUpdater) Apply() error {
	var err error
	var assetReader io.Reader
	var assetResponse *http.Response

	if assetResponse, err = http.Get(selfUpdater.assetUrl); err != nil {
		return err
	}
	defer assetResponse.Body.Close()

	if assetReader, err = selfUpdater.untarAsset(assetResponse.Body); err != nil {
		return err
	}
	return selfupdate.Apply(assetReader, selfupdate.Options{})
}

func (selfUpdater *SelfUpdater) untarAsset(assetReader io.ReadCloser) (*tar.Reader, error) {
	var err error
	var exectuablePath string
	var tarHeader *tar.Header
	var tarReader *tar.Reader
	var gzipReader *gzip.Reader

	if exectuablePath, err = os.Executable(); err != nil {
		return tarReader, err
	}
	exectuableName := filepath.Base(exectuablePath)

	if gzipReader, err = gzip.NewReader(assetReader); err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	tarReader = tar.NewReader(gzipReader)
	for {
		tarHeader, err = tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if tarHeader.Name == exectuableName {
			return tarReader, err
		}
	}
	return nil, fmt.Errorf("failed to find %s in archive", exectuableName)
}
