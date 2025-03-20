package threader_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/threatmate/threader"
)

func TestThreader(t *testing.T) {
	ctx := context.Background()

	t.Run("Default Threader", func(t *testing.T) {
		t.Run("Panic is caught", func(t *testing.T) {
			threader.Go(ctx, func() {
				panic(99)
			})
		})
	})
	t.Run("Custom Threader", func(t *testing.T) {
		t.Run("Panic", func(t *testing.T) {
			threaderInstance := threader.New()
			threaderInstance.Go(ctx, func() {
				// Empty 1
			})
			threaderInstance.Go(ctx, func() {
				// Empty 2
			})
			threaderInstance.Go(ctx, func() {
				panic(99)
			})
			err := threaderInstance.Wait()
			require.NotNil(t, err)

			err = threaderInstance.Wait()
			require.Nil(t, err)
		})
		t.Run("Error", func(t *testing.T) {
			threaderInstance := threader.New()
			threaderInstance.GoWithErr(ctx, func() error {
				return nil
			})
			threaderInstance.GoWithErr(ctx, func() error {
				return nil
			})
			threaderInstance.GoWithErr(ctx, func() error {
				return fmt.Errorf("my error")
			})
			err := threaderInstance.Wait()
			require.NotNil(t, err)

			err = threaderInstance.Wait()
			require.Nil(t, err)
		})
	})
}
