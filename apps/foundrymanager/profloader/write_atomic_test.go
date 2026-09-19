package profloader

import (
	"errors"
	"os"
	"syscall"
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

func renameErrOf(errno syscall.Errno) error {
	return &os.LinkError{Op: "rename", Old: "staged", New: "profiles.json", Err: errno}
}

func TestOverwriteMountPoint(t *testing.T) {
	t.Parallel()

	const (
		previous = "previous"
		content  = "replaced"
	)
	testCases := []struct {
		name        string
		renameErr   error
		isDirTarget bool
		want        string
		wantErr     error
	}{
		{
			name:      "busy mount point is rewritten in place",
			renameErr: renameErrOf(syscall.EBUSY),
			want:      content,
		},
		{
			name:      "other rename failure leaves the file",
			renameErr: renameErrOf(syscall.EXDEV),
			want:      previous,
			wantErr:   syscall.EXDEV,
		},
		{
			name:        "unwritable mount point",
			renameErr:   renameErrOf(syscall.EBUSY),
			isDirTarget: true,
			wantErr:     syscall.EISDIR,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := profilesPath(t)
			if testCase.isDirTarget {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
			} else {
				seedFile(t, path, previous)
			}

			err := overwriteMountPoint(path, []byte(content), testCase.renameErr)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if testCase.isDirTarget {
				return
			}
			if stored, _ := os.ReadFile(path); string(stored) != testCase.want {
				t.Errorf("content = %q, want %q", stored, testCase.want)
			}
		})
	}
}
