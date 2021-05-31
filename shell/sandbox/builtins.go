package interp

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"mvdan.cc/sh/v3/interp"
)

var builtins = map[string]func() builtinCmd{
	"ls":    func() builtinCmd { return &lsCmd{} },
	"mkdir": func() builtinCmd { return &mkdirCmd{} },
	"rm":    func() builtinCmd { return &rmCmd{} },
}

type cmdCtx struct {
	*Sandbox
	interp.HandlerContext
	runner *interp.Runner
}

// Sanitise a path within the sandbox.
func (c *cmdCtx) Sanitise(path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.Dir, path)
	}
	if !strings.HasPrefix(path, c.root) {
		return "", errors.Wrap(ErrSandboxViolation, path)
	}
	return path, nil
}

type builtinCmd interface {
	Run(bctx cmdCtx) error
}

func runBuiltinCmd(bctx cmdCtx, args []string) (error, bool) {
	factory, ok := builtins[args[0]]
	if !ok {
		return nil, false
	}
	cmd := factory()
	_, err := kong.Must(
		cmd,
		kong.Exit(func(i int) {}),
		kong.Description("List files."),
	).Parse(args[1:])
	if err != nil {
		return errors.Wrap(err, args[0]), true
	}
	return errors.Wrap(cmd.Run(bctx), args[0]), true
}
