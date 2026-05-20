package unit_test

import (
	"testing"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/cache"
	"github.com/stretchr/testify/require"
)

func TestKeyBuilder(t *testing.T) {
	kb := cache.NewKeyBuilder("fsa:")
	require.Equal(t, "fsa:entry_meta:42:7", kb.EntryMeta(42, 7))
	require.Equal(t, "fsa:fields_map:42", kb.FieldsMap(42))
	require.Equal(t, "fsa:update_types", kb.UpdateTypes())
	require.Equal(t, "fsa:custom:a:1", kb.Make("custom", "a", 1))
}

func TestKeyBuilder_EmptyPrefix(t *testing.T) {
	kb := cache.NewKeyBuilder("")
	require.Equal(t, "entry_meta:1:2", kb.EntryMeta(1, 2))
}
