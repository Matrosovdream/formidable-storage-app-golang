package unit_test

import (
	"testing"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/stretchr/testify/require"
)

func TestBuildPaginated_FirstPage(t *testing.T) {
	p := model.BuildPaginated[int]("/api/items", []int{1, 2, 3}, 10, 1, 3, 4)
	require.Equal(t, 1, p.CurrentPage)
	require.Equal(t, 1, p.From)
	require.Equal(t, 3, p.To)
	require.Equal(t, 4, p.LastPage)
	require.Equal(t, int64(10), p.Total)
	require.Nil(t, p.PrevPageURL)
	require.NotNil(t, p.NextPageURL)
	require.Equal(t, "/api/items?page=2", *p.NextPageURL)
}

func TestBuildPaginated_LastPage(t *testing.T) {
	p := model.BuildPaginated[int]("/api/items", []int{10}, 10, 4, 3, 4)
	require.Equal(t, 4, p.CurrentPage)
	require.Equal(t, 10, p.From)
	require.Equal(t, 10, p.To)
	require.Nil(t, p.NextPageURL)
	require.NotNil(t, p.PrevPageURL)
	require.Equal(t, "/api/items?page=3", *p.PrevPageURL)
}

func TestBuildPaginated_EmptyPage(t *testing.T) {
	p := model.BuildPaginated[int]("/api/items", []int{}, 0, 1, 25, 1)
	require.Equal(t, 0, p.From)
	require.Equal(t, 0, p.To)
	require.Equal(t, int64(0), p.Total)
}
