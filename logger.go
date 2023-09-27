package awos

import (
	"log/slog"
	"os"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
