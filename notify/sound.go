package notify

import (
	"io"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

type MP3Notifier struct {
	streamer beep.StreamSeekCloser
}

func NewMP3Notifier(source io.ReadCloser) (*MP3Notifier, error) {
	streamer, format, err := mp3.Decode(source)
	if err != nil {
		return nil, err
	}

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return nil, err
	}

	return &MP3Notifier{
		streamer: streamer,
	}, nil
}

func (n *MP3Notifier) Notify() error {
	done := make(chan bool)
	speaker.Play(beep.Seq(n.streamer, beep.Callback(func() {
		done <- true
	})))
	<-done

	return n.streamer.Close()
}
