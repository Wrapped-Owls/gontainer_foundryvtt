package profloader

import (
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

var fieldSuffixes = []string{ // longest-first: a shorter suffix must not shadow a longer one
	"_ADMIN_PASSWORD_SALT",
	"_MANIFEST_PATH",
	"_DATA_PATH",
	"_ADMIN_KEY",
	"_VERSION",
	"_WORLD",
	"_LABEL",
	"_NAME",
}

func FromEnv(prefix string) ([]profile.Profile, error) {
	scan := prefix + "_"
	keys := discoverKeys(scan)
	if len(keys) == 0 {
		return nil, nil
	}

	profiles := make([]profile.Profile, 0, len(keys))
	for _, key := range keys {
		pfx := scan + key
		var p profile.Profile
		if err := confloader.BindEnv(
			confloader.BindField(&p.Name, pfx+"_NAME", nil),
			confloader.BindField(&p.Label, pfx+"_LABEL", nil),
			confloader.BindField(&p.DataPath, pfx+"_DATA_PATH", nil),
			confloader.BindField(&p.AdminKey, pfx+"_ADMIN_KEY", nil),
			confloader.BindField(&p.AdminPasswordSalt, pfx+"_ADMIN_PASSWORD_SALT", nil),
			confloader.BindField(&p.Version, pfx+"_VERSION", nil),
			confloader.BindField(&p.World, pfx+"_WORLD", nil),
			confloader.BindField(&p.ManifestPath, pfx+"_MANIFEST_PATH", nil),
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func discoverKeys(scanPrefix string) []string {
	seen := make(map[string]struct{})
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, scanPrefix) {
			continue
		}
		rest := env[len(scanPrefix):]
		key := extractKey(rest)
		if key != "" {
			seen[key] = struct{}{}
		}
	}

	numeric := make([]int, 0)
	named := make([]string, 0, len(seen))
	for k := range seen {
		if n, err := strconv.Atoi(k); err == nil {
			numeric = append(numeric, n)
		} else {
			named = append(named, k)
		}
	}

	sort.Ints(numeric)
	sort.Strings(named)

	result := make([]string, 0, len(numeric)+len(named))
	for _, n := range numeric {
		result = append(result, strconv.Itoa(n))
	}
	return append(result, named...)
}

func extractKey(rest string) string {
	if eq := strings.IndexByte(rest, '='); eq >= 0 {
		rest = rest[:eq]
	}
	for _, suffix := range fieldSuffixes {
		if strings.HasSuffix(rest, suffix) {
			return rest[:len(rest)-len(suffix)]
		}
	}
	return ""
}
