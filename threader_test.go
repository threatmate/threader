package threader_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/threatmate/threader"
)

func TestThreader(t *testing.T) {
	t.Run("Default Threader", func(t *testing.T) {
		t.Run("Panic is caught", func(t *testing.T) {
			threader.DefaultThreader.Go(func() {
				panic(99)
			})
		})
	})
	t.Run("Custom Threader", func(t *testing.T) {
		t.Run("Panic", func(t *testing.T) {
			threaderInstance := threader.New()
			threaderInstance.Go(func() {
				// Empty 1
			})
			threaderInstance.Go(func() {
				// Empty 2
			})
			threaderInstance.Go(func() {
				panic(99)
			})
			err := threaderInstance.Wait()
			require.NotNil(t, err)

			err = threaderInstance.Wait()
			require.Nil(t, err)
		})
		t.Run("Error", func(t *testing.T) {
			threaderInstance := threader.New()
			threaderInstance.GoWithErr(func() error {
				return nil
			})
			threaderInstance.GoWithErr(func() error {
				return nil
			})
			threaderInstance.GoWithErr(func() error {
				return fmt.Errorf("my error")
			})
			err := threaderInstance.Wait()
			require.NotNil(t, err)

			err = threaderInstance.Wait()
			require.Nil(t, err)
		})
	})
}
