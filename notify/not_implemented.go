package notify

import (
	"context"
	"fmt"

	"github.com/kmtym1998/gh-prowl/entity"
)

type NotImplementedNotifier struct {
	name string
}

func newNotImplementedNotifier(name string) *NotImplementedNotifier {
	return &NotImplementedNotifier{name: name}
}

func (n *NotImplementedNotifier) Notify(context.Context, entity.NotificationContent) error {
	fmt.Printf("%q notifier is not implemented yet. I appreciate your contribution!\n", n.name)
	return nil
}
