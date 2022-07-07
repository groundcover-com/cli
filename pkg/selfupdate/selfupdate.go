package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/blang/semver/v4"
	"github.com/minio/selfupdate"
	"github.com/sirupsen/logrus"
	"groundcover.com/pkg/github"
	"groundcover.com/pkg/utils"
)

const (
	SKIP_SELFUPDATE_FLAG = "skip-selfupdate"
)

func TrySelfUpdate(ctx context.Context, currentVersion semver.Version) (err error) {
	var release *github.GitHubLatestRelease

	if release, err = github.NewGitHubLatestRelease(ctx); err != nil {
		return
	}
	if currentVersion.Compare(release.Version) >= 0 {
		return
	}
	shouldSelfUpdate := utils.YesNoPrompt(
		fmt.Sprintf("Your version %s is out of date! The latest version is %s.\nDo you want to update?", currentVersion, release.Version),
		true,
	)
	if !shouldSelfUpdate {
		return
	}
	return selfUpdate(ctx, release)
}

func selfUpdate(ctx context.Context, release *github.GitHubLatestRelease) (err error) {
	var tarReader *tar.Reader
	var tarHeader *tar.Header
	var gzipReader *gzip.Reader
	var assetResponse *http.Response

	if assetResponse, err = release.DownloadAsset(ctx); err != nil {
		return
	}
	defer assetResponse.Body.Close()

	if gzipReader, err = gzip.NewReader(assetResponse.Body); err != nil {
		return
	}
	defer gzipReader.Close()

	tarReader = tar.NewReader(gzipReader)
	for {
		tarHeader, err = tarReader.Next()
		if err != nil {
			return
		}
		if err == io.EOF {
			break
		}
		if tarHeader.Typeflag == tar.TypeReg {
			if err = selfupdate.Apply(tarReader, selfupdate.Options{}); err != nil {
				logrus.Error(err)
				fmt.Println("Self update has failed")
				os.Exit(1)
			}
			fmt.Println("Self update was successfully")
			os.Exit(0)
		}
	}
	return fmt.Errorf("self update has failed")
}
