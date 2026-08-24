package launcher

import "os"

// defaultBrowseDir is where a file or folder picker starts. The home
// directory is a safer starting point than the process working
// directory, which for a packaged app is somewhere the user never chose.
func defaultBrowseDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}
