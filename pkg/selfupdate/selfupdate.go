package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/github"
	"github.com/minio/selfupdate"
	"groundcover.com/pkg/ui"
)

const (
	APPLY_POLLING_RETRIES  = 1
	APPLY_POLLING_TIMEOUT  = time.Minute * 3
	APPLY_POLLING_INTERVAL = time.Second
)

var (
	devVersion = semver.MustParse("0.0.0-dev")
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

	return selfUpdater, nil
}

func (selfUpdater *SelfUpdater) fetchVersion(githubRelease *github.RepositoryRelease) error {
	var err error
	var version semver.Version

	if version, err = semver.ParseTolerant(githubRelease.GetTagName()); err != nil {
		return err
	}

	selfUpdater.Version = version
	return nil
}

func (selfUpdater *SelfUpdater) fetchAsset(ctx context.Context, client *github.Client, githubRelease *github.RepositoryRelease) error {
	var err error
	var assetUrl string

	assetSuffix := fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	for _, asset := range githubRelease.Assets {
		if strings.HasSuffix(asset.GetName(), assetSuffix) {
			selfUpdater.assetId = asset.GetID()
			if _, assetUrl, err = client.Repositories.DownloadReleaseAsset(ctx, selfUpdater.githubOwner, selfUpdater.githubRepo, selfUpdater.assetId); err != nil {
				return err
			}
			selfUpdater.assetUrl = assetUrl
			return nil
		}
	}

	return fmt.Errorf("failed to find asset for %s", assetSuffix)
}

func (selfUpdater *SelfUpdater) IsLatestNewer(currentVersion semver.Version) bool {
	return selfUpdater.Version.GT(currentVersion)
}

func (selfUpdater *SelfUpdater) IsDevVersion(currentVersion semver.Version) bool {
	return currentVersion.Equals(devVersion)
}

func (selfUpdater *SelfUpdater) Apply(ctx context.Context) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner(fmt.Sprintf("Downloading cli version: %s", selfUpdater.Version))
	spinner.StopMessage("cli update was successfully")
	spinner.StopFailMessage("cli update has failed")

	spinner.Start()
	defer spinner.Stop()

	err = spinner.Poll(ctx, selfUpdater.apply, APPLY_POLLING_INTERVAL, APPLY_POLLING_TIMEOUT, APPLY_POLLING_RETRIES)

	if err == nil {
		return nil
	}

	spinner.StopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return errors.New("timeout waiting for cli download")
	}

	return err
}

func (selfUpdater *SelfUpdater) apply() error {
	var err error

	var assetResponse *http.Response
	if assetResponse, err = http.Get(selfUpdater.assetUrl); err != nil {
		return ui.RetryableError(err)
	}
	defer assetResponse.Body.Close()

	var assetReader io.Reader
	if assetReader, err = selfUpdater.untarAsset(assetResponse.Body); err != nil {
		return ui.RetryableError(err)
	}

	if err = selfupdate.Apply(assetReader, selfupdate.Options{}); err != nil {
		return ui.RetryableError(err)
	}

	return nil
}

func (selfUpdater *SelfUpdater) untarAsset(assetReader io.ReadCloser) (*tar.Reader, error) {
	var err error
	var exectuablePath string
	var tarHeader *tar.Header
	var tarReader *tar.Reader
	var gzipReader *gzip.Reader

	if exectuablePath, err = os.Executable(); err != nil {
		return nil, err
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
			return tarReader, nil
		}
	}

	return nil, fmt.Errorf("failed to find %s in archive", exectuableName)
}
