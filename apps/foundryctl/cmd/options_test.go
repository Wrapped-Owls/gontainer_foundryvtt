package cmd

import (
	"log/slog"
	"testing"
)

func TestOptionsReturnsInt(t *testing.T) {
	_ = Options(nil, slog.Default())
}
