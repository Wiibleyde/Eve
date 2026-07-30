package api

import (
	"net/http"
	"reflect"
	"slices"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

const (
	specPath  = "/openapi.yaml"
	uiPath    = "docs"
	specTitle = "Eve API"
)

const specDescription = "Read-only HTTP API exposed by the Eve Discord bot. " +
	"Every endpoint is a GET and requires no authentication. " +
	"CORS is open to all origins but restricted to the GET method."

type statsQuery struct {
	Game  string `query:"game" description:"Restrict to a single game ID. Overrides the default active-only filter."`
	All   string `query:"all" description:"Set to true to include inactive games. Ignored when game is given." enum:"true"`
	Guild string `query:"guild" description:"Restrict to a single Discord guild snowflake ID." example:"1091918939407908944"`
}

type winnersQuery struct {
	Game  string `query:"game" description:"Restrict to a single game ID."`
	Guild string `query:"guild" description:"Restrict to a single Discord guild snowflake ID." example:"1091918939407908944"`
	Limit int    `query:"limit" description:"Maximum number of winners to return. Values above 500 are clamped to 500." minimum:"1" maximum:"500" default:"500"`
}

func unprefixedDefName(t reflect.Type, defaultDefName string) string {
	if name := t.Name(); name != "" {
		return name
	}
	return defaultDefName
}

func requireResponseFields(p jsonschema.InterceptPropParams) error {
	if !p.Processed || p.ParentSchema == nil {
		return nil
	}

	oc, ok := openapi.OperationCtx(p.Context)
	if !ok || !oc.IsProcessingResponse() {
		return nil
	}

	if !slices.Contains(p.ParentSchema.Required, p.Name) {
		p.ParentSchema.Required = append(p.ParentSchema.Required, p.Name)
	}
	return nil
}

func buildSpec(version string, endpoints []apiRoute) ([]byte, error) {
	reflector := openapi3.NewReflector()
	reflector.DefaultOptions = append(reflector.DefaultOptions,
		jsonschema.InterceptDefName(unprefixedDefName),
		jsonschema.InterceptProp(requireResponseFields),
	)

	reflector.Spec.Openapi = "3.0.3"
	reflector.Spec.Info.
		WithTitle(specTitle).
		WithVersion(version).
		WithDescription(specDescription)
	reflector.Spec.WithServers(openapi3.Server{URL: "/"})
	reflector.Spec.WithSecurity(map[string][]string{})

	for _, route := range endpoints {
		oc, err := reflector.NewOperationContext(http.MethodGet, route.path)
		if err != nil {
			return nil, err
		}

		oc.SetTags(route.tag)
		oc.SetID(route.id)
		oc.SetSummary(route.summary)

		if route.query != nil {
			oc.AddReqStructure(route.query)
		}

		for _, res := range route.responses {
			oc.AddRespStructure(res.body, func(cu *openapi.ContentUnit) {
				cu.HTTPStatus = res.status
				cu.Description = res.description
			})
		}

		if err := reflector.AddOperation(oc); err != nil {
			return nil, err
		}
	}

	return reflector.Spec.MarshalYAML()
}
