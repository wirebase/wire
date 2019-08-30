package bundle

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/advanderveer/9d587b277/line/vfsgen"
)

// Bundle describes the directory in which all static assets to eventually
// generate an embeddable filesystem
type Bundle struct {
	dir string
}

// New creates a new bundle
func New() (b *Bundle, err error) {
	b = &Bundle{}
	b.dir, err = ioutil.TempDir("", "bundle_")
	if err != nil {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	return
}

// Dir returns the directory this bundle resides or resided in
func (b *Bundle) Dir() string { return b.dir }

// Clear will remove the bundle directory and all assets in it
func (b *Bundle) Clear() (err error) {
	err = os.RemoveAll(b.dir)
	if err != nil {
		return fmt.Errorf("failed to remove bundle dir: %w", err)
	}

	return
}

// Write the bundle as an go file that embeds the assets in the bundle
func (b *Bundle) Write(o string) error {
	fs := http.Dir(b.dir)
	if err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:  o,
		BuildTags: "!wasm",
	}); err != nil {
		return fmt.Errorf("failed to generate embed file: %w", err)
	}

	return nil
}
