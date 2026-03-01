package main

import (
	"context"
	"fmt"

	"phant/internal/dump"
)

const DumpEventSchemaVersion = dump.SchemaVersion

var ErrUnsupportedSchemaVersion = dump.ErrUnsupportedSchemaVersion

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SupportedDumpEventSchemaVersion() int {
	return dump.SchemaVersion
}

func (a *App) DecodeDumpEventNDJSONLine(line string) (*dump.Event, error) {
	return dump.DecodeNDJSONLine(line)
}
