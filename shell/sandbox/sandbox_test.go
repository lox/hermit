package interp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	root := t.TempDir()
	sandbox, err := New(root)
	require.NoError(t, err)
	err = sandbox.Eval(`
		set -e
		echo > .t
		echo "Hello world" > t
		mkdir test
		echo hi > test/foo
		(cd test && ls -l)
		ls
		ls -la
		ls -l t
		rm -rf *
		echo ../../*
		echo empty
		ls -l
		cd /
	`)
	require.NoError(t, err)
}
