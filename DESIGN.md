# Design
By default it will detect the current path and the package name if there are
any go files. If it's not main show a warning that the server is only usefull
in directories that can build a non-library function

If `GOOS=js GOARCH=wasm go list -f '{{.GoFiles}}' .` returns at least one file
the command will start start compiling wasm with wasm_exec.js polyfills etc and add it
to the bundle.

If the bundler sees/has at least one file (maybe wasm file) It will start producing
an embed file for the detected package.

If `go list -f '{{.GoFiles}}' .` returns at least one file it will start re-compiling
to a temporary binary and starting it. If the binary exits too quickly with a non-zero
exit code do not restart it until something changed

# User Story
If there is nothing to Go compile in the current directory for the js/wasm target
with a main function it will start generating `assets.go` in the current dir.

Whenever there is a main function that compiles in the !js build tag we will
do this automatically to temporary binary file. If there are no buildable sources
or no main we do nothing and give instructions on how to do it.

Now show a main function with the !wasm build tag that serves just the fileserver;
with no root component using our handler. leave that one empty. Editing anything
should rebuild the filesystem, and re-start the server.

Now show how to create the first component that just shows a basic html page. Show
how to update the server to render this component instead of just files. Note
that this just takes a component approach to what would normally something with
html templating.

Now add a main function with the "wasm" build tag that just runs hello, world. Note
how the development server detects this automatically and will start building the
wasm binary. Show how to edit the html to load the wasm and display the hello
world in the console. Note that wasm functionality is optional and that components
can also be just rendered for static websites or server side html only.

Create another component called 'counter' and embed it into the 'doc' component.
Note how the default html of this component without any interaction will render
as we expect it. This is usefull for SEO and will provide the user with feedback
while the wasm is downloading if they are on a slow connection.

Now edit the wasm main to mount the app component to make the page interactive.
Conclude and explain what steps happened in the background:
- It creates a temporary bundle directory
- It creates a temporary embed file
- It creates a temporary server file
- It compiles the temporary server to a temporary binary
- It manages the execution of this binary

## Resources
- Conditional Compilation, and using go list: https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool

## Future Ideas
- [ ] Turn the process into an http proxy that shows compile errors and other info
      visually (dogfooding our own framework)
- [ ] Allow arbitrary code to operate on the bundle dir before it is turned into
      the embed file: i.e minification
- [ ] Allow arbitrary code to be run before each restart; i.e fixture data
- [ ] Allow for rendering static html files to the bundle based on the component
- [ ] Add an exponential backoff mechanic to the polling such that when
  nothing changed for a bit the interval is increased
- [ ] If scanning takes a long time, increase the scanning interval
- [ ] Optimize scanning by starting with directories that are have
  changed recently (priority queue)
- [ ] When scanning visit (large) directories less often, or let them have
  a separate timer
- [ ] Detect loops, where detecting a changed file causes a change to happen