package lifecycle

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
)

// WriteAdminPassword writes the PBKDF2-hashed admin key to
// <dataPath>/Config/admin.txt. When plaintext is empty, any existing
// admin.txt is removed (clearing the admin password).
//
// Returns true iff the file was created/updated/removed.
func WriteAdminPassword(dataPath, plaintext, salt string) (bool, error) {
	dest := filepath.Join(ConfigDir(dataPath), "admin.txt")
	if strings.TrimSpace(plaintext) == "" {
		err := os.Remove(dest)
		if err == nil {
			return true, nil
		}
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	hashed, err := config.HashAdminKey(plaintext, salt)
	if err != nil {
		return false, err
	}
	if err = os.MkdirAll(filepath.Dir(dest), fsperm.Dir); err != nil {
		return false, err
	}
	if existing, err := os.ReadFile(dest); err == nil && string(existing) == hashed {
		return false, nil
	}
	if err = os.WriteFile(dest, []byte(hashed), fsperm.Secret); err != nil {
		return false, err
	}
	return true, nil
}
