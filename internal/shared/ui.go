package shared

import (
	"context"
	"fmt"

	cview "code.rocketnine.space/tslocum/cview"
	fluffycore_async "github.com/fluffy-bunny/fluffycore/async"
	async "github.com/reugn/async"
	"github.com/rs/zerolog"
)

type (
	UI struct {
		Main   *cview.TextView
		Footer *cview.TextView
		App    *cview.Application
		Future async.Future[*fluffycore_async.AsyncResponse]
	}
)

func NewUI(ctx context.Context) *UI {
	log := zerolog.Ctx(ctx).With().Str("command", "NewUI").Logger()
	app := cview.NewApplication()
	app.EnableMouse(true)
	newPrimitive := func(text string) *cview.TextView {
		textView := cview.NewTextView()
		textView.SetDynamicColors(true)
		textView.SetRegions(true)
		textView.SetWordWrap(true)
		textView.SetChangedFunc(func() {
			app.Draw()
		})
		fmt.Fprintf(textView, "%s\n", text)
		return textView
	}
	main := newPrimitive("")
	footer := newPrimitive("")

	grid := cview.NewGrid()
	grid.SetRows(3, 0, 3)
	grid.SetColumns(30, 0, 30)
	grid.SetBorders(true)
	//	grid.AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false)

	grid.AddItem(footer, 2, 0, 1, 3, 0, 0, false)
	grid.AddItem(main, 1, 0, 1, 3, 0, 0, false)

	app.SetRoot(grid, true)

	future := fluffycore_async.ExecuteWithPromiseAsync(func(promise async.Promise[*fluffycore_async.AsyncResponse]) {
		var err error
		defer func() {
			promise.Success(&fluffycore_async.AsyncResponse{
				Message: "End Serve - tview",
				Error:   err,
			})
		}()

		err = app.Run()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to run app")
		}

	})
	ui := &UI{
		Main:   main,
		Footer: footer,
		App:    app,
		Future: future,
	}
	return ui
}
