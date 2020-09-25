package config

import (
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/worksinmagic/ytfeed/mock"
)

func TestConfig(t *testing.T) {
	var cfg *Configuration

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mock.NewMockValidator(ctrl)

	t.Run("Validate success", func(t *testing.T) {
		version := "v100.0.0"
		os.Setenv("YTFEED_VERSION", version)

		cfg = New()
		require.NotNil(t, cfg)
		cfg.StorageBackend = StorageBackendDisk
		cfg.DiskDirectory = "/"

		cfg.validator = mockValidator
		mockValidator.EXPECT().Struct(gomock.AssignableToTypeOf(cfg)).Return(nil)

		err := cfg.Validate()
		require.NoError(t, err)
		require.Equal(t, version, cfg.Version)
		require.Equal(t, DefaultHost, cfg.Host)
	})

	t.Run("Validate failed", func(t *testing.T) {
		cfg = New()
		require.NotNil(t, cfg)
		cfg.StorageBackend = StorageBackendDisk
		cfg.DiskDirectory = "/"

		cfg.validator = mockValidator
		mockValidator.EXPECT().Struct(gomock.AssignableToTypeOf(cfg)).Return(errors.New("error"))

		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("Validate no storage backend", func(t *testing.T) {
		cfg = New()
		require.NotNil(t, cfg)

		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("Validate failed disk", func(t *testing.T) {
		cfg = New()
		require.NotNil(t, cfg)
		cfg.StorageBackend = StorageBackendDisk

		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("Validate failed S3", func(t *testing.T) {
		cfg = New()
		require.NotNil(t, cfg)
		cfg.StorageBackend = StorageBackendS3

		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("Validate failed GCS", func(t *testing.T) {
		cfg = New()
		require.NotNil(t, cfg)
		cfg.StorageBackend = StorageBackendGCS

		err := cfg.Validate()
		require.Error(t, err)
	})
}
