package notify

import (
	"context"

	"github.com/kmtym1998/gh-prowl/entity"
)

type noopNotifier struct{}

func NewNoopNotifier() *noopNotifier {
	return &noopNotifier{}
}

func (n *noopNotifier) Notify(context.Context, entity.NotificationContent) error {
	return nil
}
