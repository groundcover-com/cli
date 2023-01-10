package utils

import (
	"github.com/skratchdot/open-golang/open"
	"groundcover.com/pkg/ui"
)

func TryOpenBrowser(message string, url string) {
	ui.GlobalWriter.Printf("%s %s\n", message, ui.GlobalWriter.UrlLink(url))
	open.Run(url)
}
