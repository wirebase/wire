package project

import (
	"fmt"
	"io"
)

// UI provides feedback to the user
type UI interface {
	ShowRebuildStarted()
	ShowRebuildDone()
	ShowConfigLoaded()
	ShowBundlingDone()
	ShowRunningDone()
	ShowBundleCreated()
	ShowWasmBundled()
	ShowEmbedFileWritten()
	ShowBuildingDone()
}

// TerseTerminal is a ui implementation that writes a terse output of the
// building process to the terminal
type TerseTerminal struct{ w io.Writer }

// NewTerseTerminal returns a terse terminal ui
func NewTerseTerminal(w io.Writer) (ui *TerseTerminal) {
	ui = &TerseTerminal{w}
	return
}

// ShowRebuildStarted is called when the build starts
func (ui *TerseTerminal) ShowRebuildStarted() { fmt.Fprintf(ui.w, "rebuilding") }

// ShowRebuildDone is called when the build is done
func (ui *TerseTerminal) ShowRebuildDone() { fmt.Fprintf(ui.w, "done\n") }

// ShowConfigLoaded is called when the config is (re)loaded
func (ui *TerseTerminal) ShowConfigLoaded() { fmt.Fprintf(ui.w, ".") }

// ShowBundlingDone is called when the bundling is done
func (ui *TerseTerminal) ShowBundlingDone() { fmt.Fprintf(ui.w, ".") }

// ShowRunningDone is called when process run is done
func (ui *TerseTerminal) ShowRunningDone() { fmt.Fprintf(ui.w, ".") }

// ShowBundleCreated is called when the bundle is created
func (ui *TerseTerminal) ShowBundleCreated() { fmt.Fprintf(ui.w, ".") }

// ShowWasmBundled is called wehn the wasm has been bundled
func (ui *TerseTerminal) ShowWasmBundled() { fmt.Fprintf(ui.w, ".") }

// ShowEmbedFileWritten is called when the embed file was written to disk
func (ui *TerseTerminal) ShowEmbedFileWritten() { fmt.Fprintf(ui.w, ".") }

// ShowBuildingDone is called when the binary was built
func (ui *TerseTerminal) ShowBuildingDone() { fmt.Fprintf(ui.w, ".") }
