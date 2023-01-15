package utils

import (
	"github.com/skratchdot/open-golang/open"
	"groundcover.com/pkg/ui"
)

func TryOpenBrowser(writer *ui.Writer, message string, url string) {
	writer.PrintUrl(message, url)
	open.Run(url)
}
