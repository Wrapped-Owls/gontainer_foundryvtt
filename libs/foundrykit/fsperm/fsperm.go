package fsperm

import "io/fs"

const (
	Dir    fs.FileMode = 0o755
	File   fs.FileMode = 0o644
	Secret fs.FileMode = 0o600
)
