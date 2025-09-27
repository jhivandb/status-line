package sections

import (
	"fmt"
	"os"
	"strings"

	"github.com/jhivandb/status-line/internal/api"
)

type Path struct {
	config string
	In     api.InputData
}

func (p *Path) Render() string {

	return fmt.Sprintf("%s%s", api.ColorBlue, getPath(p.In))
}

func getPath(in api.InputData) string {
	workDir := in.CWD

	homeDir, err := os.UserHomeDir()

	if err == nil && strings.HasPrefix(workDir, homeDir) {
		workDir = "\uf07b ~" + workDir[len(homeDir):]
	}

	return workDir
}
