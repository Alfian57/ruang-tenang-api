package docs

import "testing"

func TestSwaggerInfoBasics(t *testing.T) {
	if SwaggerInfo == nil {
		t.Fatalf("expected swagger info")
	}
	if SwaggerInfo.InstanceName() == "" {
		t.Fatalf("expected swagger instance name")
	}
	doc := SwaggerInfo.ReadDoc()
	if len(doc) == 0 {
		t.Fatalf("expected swagger doc content")
	}
}
