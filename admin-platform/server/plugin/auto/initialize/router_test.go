package initialize

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitializeRouterRegistersOnlyAutoCompatibilityPrefixes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	InitializeRouter(engine)

	foundAutoCode := false
	for _, route := range engine.Routes() {
		if route.Path == "/autoCode/getDB" {
			foundAutoCode = true
		}
		if route.Path == "/skills/getTools" {
			t.Fatalf("auto plugin should not register /skills/getTools; the AI plugin owns skills routes")
		}
	}

	if !foundAutoCode {
		t.Fatalf("expected /autoCode/getDB to be registered by plugin router")
	}
}
