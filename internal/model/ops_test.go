package model_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model"
)

type emptyModel struct {
	Name string `json:"name"`
}

func (d emptyModel) GetName() string {
	return d.Name
}

func TestModel(t *testing.T) {
	t.Run("nil storage", func(t *testing.T) {
		var err error

		require.ErrorIs(t, model.Delete(t.Context(), nil, "path"), errs.ErrNilValue)
		require.ErrorIs(t, model.Save(t.Context(), nil, "path", nil), errs.ErrNilValue)

		_, err = model.List(t.Context(), nil, "path")
		require.ErrorIs(t, err, errs.ErrNilValue)

		_, err = model.Get[any](t.Context(), nil, "test")
		require.ErrorIs(t, err, errs.ErrNilValue)
	})

	t.Run("ops", func(t *testing.T) {
		var path = "path/test"
		storage := &logical.InmemStorage{}

		data, err := model.Get[emptyModel](t.Context(), storage, path)
		require.NoError(t, err)
		require.Nil(t, data)

		data = &emptyModel{Name: "test"}

		require.NoError(t, model.Save(t.Context(), storage, "path", data))

		data, err = model.Get[emptyModel](t.Context(), storage, path)
		require.NoError(t, err)
		require.NotNil(t, data)
		require.Equal(t, "test", data.GetName())

		var entries []string
		entries, err = model.List(t.Context(), storage, "path/")
		require.Len(t, entries, 1)
		require.NoError(t, err)

		require.NoError(t, model.Delete(t.Context(), storage, path))

		entries, err = model.List(t.Context(), storage, "path/")
		require.Len(t, entries, 0)
		require.NoError(t, err)
	})
}
