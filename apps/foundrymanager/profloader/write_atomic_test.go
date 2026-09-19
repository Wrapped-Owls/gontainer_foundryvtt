package profloader

import (
	"os"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

func TestWriteFileAtomic(t *testing.T) {
	t.Parallel()

	path := profilesPath(t)
	seedFile(t, path, "previous")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if err = writeFileAtomic(path, []byte("replaced")); err != nil {
		t.Fatalf("write: %v", err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if os.SameFile(before, after) {
		t.Error("the file was truncated in place instead of replaced via rename")
	}
	if after.Mode().Perm() != fsperm.Secret {
		t.Errorf("perm = %v, want %v", after.Mode().Perm(), fsperm.Secret)
	}
}
