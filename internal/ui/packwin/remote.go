package packwin

import (
	"github.com/PalisadeMC/Packwiz-Studio/internal/forge"
)

// remoteWebURL turns a git remote URL into the repository's page, so the
// open remote action works for SSH remotes too.
func remoteWebURL(raw string) (string, error) {
	remote, err := forge.ParseRemote(raw)
	if err != nil {
		return "", err
	}
	return remote.WebURL(), nil
}
