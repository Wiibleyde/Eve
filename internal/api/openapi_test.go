package api

import (
	"strings"
	"testing"

	"Eve/internal/api/controllers"
)

func TestBuildSpecCoversEveryRoute(t *testing.T) {
	endpoints := routes(controllers.StatusController{}, controllers.LotoController{})
	if len(endpoints) == 0 {
		t.Fatal("no routes declared")
	}

	spec, err := buildSpec("test", endpoints)
	if err != nil {
		t.Fatalf("buildSpec returned error: %v", err)
	}

	body := string(spec)
	for _, want := range []string{"openapi: 3.0.3", "version: test"} {
		if !strings.Contains(body, want) {
			t.Errorf("spec is missing %q", want)
		}
	}

	for _, route := range endpoints {
		if !strings.Contains(body, route.path+":") {
			t.Errorf("spec is missing path %q", route.path)
		}
		if !strings.Contains(body, "operationId: "+route.id) {
			t.Errorf("spec is missing operationId %q", route.id)
		}
		if route.handler == nil {
			t.Errorf("route %q has no handler", route.path)
		}
		if route.summary == "" {
			t.Errorf("route %q has no summary", route.path)
		}
		if len(route.responses) == 0 {
			t.Errorf("route %q declares no responses", route.path)
		}
		for _, res := range route.responses {
			if res.description == "" {
				t.Errorf("route %q has response %d without description", route.path, res.status)
			}
		}
	}
}
