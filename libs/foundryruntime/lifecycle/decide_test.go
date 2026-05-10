package lifecycle

import "testing"

func TestDecideInstall(t *testing.T) {
	cases := []struct {
		name    string
		info    InstalledInfo
		desired string
		want    InstallAction
	}{
		{"fresh", InstalledInfo{}, "12.331", ActionInstall},
		{"fresh empty desired", InstalledInfo{}, "", ActionInstall},
		{"match", InstalledInfo{Present: true, Version: "12.331"}, "12.331", ActionNone},
		{"upgrade", InstalledInfo{Present: true, Version: "11.315"}, "12.331", ActionUpgrade},
		{"unknown version, no desired", InstalledInfo{Present: true}, "", ActionNone},
		{"unknown version, desired set", InstalledInfo{Present: true}, "12.331", ActionNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DecideInstall(tc.info, tc.desired)
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestInstallActionString(t *testing.T) {
	cases := []struct {
		a    InstallAction
		want string
	}{
		{ActionNone, "none"},
		{ActionInstall, "install"},
		{ActionUpgrade, "upgrade"},
		{InstallAction(99), "action(99)"},
	}
	for _, tc := range cases {
		if got := tc.a.String(); got != tc.want {
			t.Errorf("InstallAction(%d).String() = %q, want %q", int(tc.a), got, tc.want)
		}
	}
}
