package utils

import (
	"github.com/skratchdot/open-golang/open"
	"groundcover.com/pkg/ui"
)

func TryOpenBrowser(message string, url string) {
	ui.SingletonWriter.Printf("%s %s\n", message, ui.SingletonWriter.UrlLink(url))
	open.Run(url)
}
