package notify

import (
	"io"
	"runtime"

	"github.com/kmtym1998/gh-prowl/entity"
)

// NewNotifier creates entity.Notifier depending on the
func NewNotifier(source io.ReadCloser) (entity.Notifier, error) {
	goos := runtime.GOOS
	switch goos {
	case "darwin":
		return newDarwinNotifier(source)
	default:
		return newNotImplementedNotifier(goos), nil
	}
}
