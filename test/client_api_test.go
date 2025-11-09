package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanta7/hydros/internal/client"
)

func Test_Client_Create(t *testing.T) {
	app := SetupTestApp(t)
	defer app.Cleanup()

	err := app.ClientUC.CreateClient(ctx, &client.Client{
		Name:   "test",
		Secret: "test",
	})
	assert.NoError(t, err)
}
