package dang

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstructorHook(t *testing.T) {
	for _, source := range []string{
		`type Item { pub name: String! }
let item = Item(name: "original")`,
		`type Item {
  pub name: String!
  new(name: String!) { self.name = name; self }
}
let item = Item(name: "original")`,
	} {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.dang"), []byte(source), 0600))
		calls := 0
		ctx := WithConstructorHook(t.Context(), func(ctx context.Context, scope ValueScope, obj *Object) (*Object, error) {
			if obj.Mod.Name() != "Item" {
				return obj, nil
			}
			calls++
			name, found, err := obj.Lookup(ctx, "name")
			require.NoError(t, err)
			require.True(t, found)
			require.Equal(t, StringValue{Val: "original"}, name)
			obj.Bind("name", StringValue{Val: "filtered"}, PublicVisibility)
			return obj, nil
		})
		env, err := RunDir(ctx, dir, false)
		require.NoError(t, err)
		item, found, err := env.Lookup(ctx, "item")
		require.NoError(t, err)
		require.True(t, found)
		name, _, err := item.(*Object).Lookup(ctx, "name")
		require.NoError(t, err)
		require.Equal(t, StringValue{Val: "filtered"}, name)
		require.Equal(t, 1, calls)

		// The hook is scoped to this evaluation; ordinary construction is unchanged.
		env, err = RunDir(t.Context(), dir, false)
		require.NoError(t, err)
		item, _, err = env.Lookup(t.Context(), "item")
		require.NoError(t, err)
		name, _, err = item.(*Object).Lookup(t.Context(), "name")
		require.NoError(t, err)
		require.Equal(t, StringValue{Val: "original"}, name)
		ctx = WithConstructorHook(t.Context(), func(context.Context, ValueScope, *Object) (*Object, error) {
			return nil, errors.New("hook failure")
		})
		_, err = RunDir(ctx, dir, false)
		require.ErrorContains(t, err, "hook failure")
	}
}
