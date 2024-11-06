package main

import (
	"context"

	"github.com/WM1rr0rB8/librariesTest/backend/golang/logging"

	"software_test/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("can't init a new app", logging.ErrAttr(err))
		return
	}

	err = newApp.Run(ctx)
	if err != nil {
		logging.L(ctx).Error("can't run app work ground", logging.ErrAttr(err))
	}
}
