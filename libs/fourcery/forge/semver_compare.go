package forge

import "github.com/Masterminds/semver/v3"

func compareSemver(a, b string) int {
	av, errA := semver.NewVersion(a)
	bv, errB := semver.NewVersion(b)
	switch {
	case errA != nil && errB != nil:
		return 0
	case errA != nil:
		return -1
	case errB != nil:
		return 1
	}
	return av.Compare(bv)
}
