package action

import (
	"context"
	"os"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

type fileReplaceRunner struct{}

func FileReplace() Runner { return fileReplaceRunner{} }

func (fileReplaceRunner) Run(_ context.Context, act manifest.Action, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), fsperm.Dir); err != nil {
		return err
	}
	return os.WriteFile(dest, []byte(act.Content), fsperm.File)
}
