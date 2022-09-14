package helm

import (
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func (helmClient *Client) AddRepo(name, url string) error {
	var err error

	repoEntry := &repo.Entry{
		URL:  url,
		Name: name,
	}

	var chartRepo *repo.ChartRepository
	if chartRepo, err = repo.NewChartRepository(repoEntry, getter.All(helmClient.settings)); err != nil {
		return err
	}

	if _, err = chartRepo.DownloadIndexFile(); err != nil {
		return err
	}

	repoFile := repo.NewFile()
	repoFile.Add(repoEntry)

	return repoFile.WriteFile(helmClient.settings.RepositoryConfig, 0755)
}
