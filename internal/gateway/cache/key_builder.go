package cache

import (
	"fmt"
	"strings"
)

type KeyBuilder struct {
	prefix string
}

func NewKeyBuilder(prefix string) *KeyBuilder {
	return &KeyBuilder{prefix: prefix}
}

func (b *KeyBuilder) Make(namespace string, parts ...any) string {
	var sb strings.Builder
	sb.WriteString(b.prefix)
	sb.WriteString(namespace)
	for _, p := range parts {
		sb.WriteString(":")
		sb.WriteString(fmt.Sprint(p))
	}
	return sb.String()
}

func (b *KeyBuilder) EntryMeta(siteID, entryID int64) string {
	return b.Make("entry_meta", siteID, entryID)
}

func (b *KeyBuilder) FieldsMap(siteID int64) string {
	return b.Make("fields_map", siteID)
}

func (b *KeyBuilder) UpdateTypes() string {
	return b.Make("update_types")
}
