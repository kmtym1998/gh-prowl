package notify

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/kmtym1998/gh-prowl/entity"
)

type darwinNotifier struct {
	sourceFileLocation string
}

func newDarwinNotifier(source io.ReadCloser) (*darwinNotifier, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	soundFileLocation := filepath.Join(home, ".gh-prowl", "chime.mp3")

	if _, err := os.ReadFile(soundFileLocation); err == nil {
		return &darwinNotifier{
			sourceFileLocation: soundFileLocation,
		}, nil
	}

	if err := os.MkdirAll(filepath.Dir(soundFileLocation), 0755); err != nil {
		return nil, err
	}

	soundFile, err := os.Create(soundFileLocation)
	if err != nil {
		return nil, err
	}
	defer soundFile.Close()

	if _, err := io.Copy(soundFile, source); err != nil {
		return nil, err
	}

	return &darwinNotifier{
		sourceFileLocation: soundFileLocation,
	}, nil
}

func (n *darwinNotifier) Notify(ctx context.Context, content entity.NotificationContent) error {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		if err := exec.CommandContext(ctx, "afplay", n.sourceFileLocation).Run(); err != nil {
			color.Red("failed to play sound: %v\n", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		osascriptArg := fmt.Sprintf(`display notification %q with title %q`, content.Message, content.Title)
		if err := exec.CommandContext(ctx, "osascript", "-e", osascriptArg).Run(); err != nil {
			color.Red("failed to display notification: %v\n", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}
