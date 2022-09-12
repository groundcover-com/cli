package utils

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
	"groundcover.com/pkg/ui"
)

func TryOpenBrowser(message string, url string) {
	fmt.Printf("%s %s\n", message, ui.UrlLink(url))
	open.Run(url)
}
