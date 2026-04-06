package routes

import (
	"testing"

	"github.com/gofiber/fiber/v3"
)

func dummyHandler(c fiber.Ctx) error {
	return c.SendString("ok")
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(r.validators) == 0 {
		t.Error("expected default validators to be registered")
	}
}

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	group := &RouteGroup{
		Name: "test",
		Routes: []Route{
			{Method: "GET", Path: "/test", Handler: dummyHandler},
		},
	}

	if err := r.Register(group); err != nil {
		t.Errorf("Register() error = %v", err)
	}

	if len(r.groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(r.groups))
	}
}

func TestRegistryRegisterNil(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(nil); err == nil {
		t.Error("expected error for nil group")
	}
}

func TestRegistryMustRegister(t *testing.T) {
	r := NewRegistry()

	defer func() {
		if recover() == nil {
			t.Error("expected panic for invalid group")
		}
	}()

	r.MustRegister(&RouteGroup{})
}

func TestRegistryAudit(t *testing.T) {
	r := NewRegistry()

	group := &RouteGroup{
		Name:   "test",
		Prefix: "/api",
		Middlewares: []Middleware{
			{Name: "auth", Handler: dummyHandler},
		},
		Routes: []Route{
			{Method: "GET", Path: "/users", Summary: "List users", Auth: AuthRequired, Handler: dummyHandler},
		},
		SubGroups: []*RouteGroup{
			{
				Name:   "admin",
				Prefix: "/admin",
				Routes: []Route{
					{Method: "GET", Path: "/settings", Summary: "Get settings", Auth: AuthServiceKey, Internal: true, Handler: dummyHandler},
				},
			},
		},
	}

	if err := r.Register(group); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	entries := r.Audit()
	if len(entries) != 2 {
		t.Errorf("expected 2 audit entries, got %d", len(entries))
	}
}

func TestValidateAuthConsistency(t *testing.T) {
	tests := []struct {
		name    string
		route   Route
		wantErr bool
	}{
		{"consistent auth", Route{Auth: AuthRequired, Roles: []string{"admin"}, Handler: dummyHandler}, false},
		{"roles without auth", Route{Auth: AuthNone, Roles: []string{"admin"}, Handler: dummyHandler}, true},
		{"scopes without auth", Route{Auth: AuthNone, Scopes: []string{"read"}, Handler: dummyHandler}, true},
		{"public with required auth", Route{Auth: AuthRequired, Public: true, Handler: dummyHandler}, true},
		{"public with optional auth", Route{Auth: AuthOptional, Public: true, Handler: dummyHandler}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &RouteGroup{Name: "test"}
			err := ValidateAuthConsistency(group, tt.route, "/test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthConsistency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePublicRoutes(t *testing.T) {
	tests := []struct {
		name    string
		route   Route
		wantErr bool
	}{
		{"public with summary", Route{Public: true, Auth: AuthNone, Summary: "Health check endpoint", Handler: dummyHandler}, false},
		{"public without summary", Route{Public: true, Auth: AuthNone, Handler: dummyHandler}, true},
		{"non-public", Route{Auth: AuthRequired, Handler: dummyHandler}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &RouteGroup{Name: "test"}
			err := ValidatePublicRoutes(group, tt.route, "/test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePublicRoutes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMiddlewareDependencies(t *testing.T) {
	tests := []struct {
		name    string
		group   *RouteGroup
		route   Route
		wantErr bool
	}{
		{
			name:    "no dependencies",
			group:   &RouteGroup{Name: "test"},
			route:   Route{Handler: dummyHandler},
			wantErr: false,
		},
		{
			name: "satisfied dependency",
			group: &RouteGroup{
				Name: "test",
				Middlewares: []Middleware{
					{Name: "csrf"},
					{Name: "auth", DependsOn: []string{"csrf"}},
				},
			},
			route:   Route{Handler: dummyHandler},
			wantErr: false,
		},
		{
			name: "unsatisfied dependency",
			group: &RouteGroup{
				Name: "test",
				Middlewares: []Middleware{
					{Name: "auth", DependsOn: []string{"csrf"}},
				},
			},
			route:   Route{Handler: dummyHandler},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMiddlewareDependencies(tt.group, tt.route, "/test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMiddlewareDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRouteValidation(t *testing.T) {
	r := Route{Method: "", Path: "/test", Handler: nil}
	if err := r.Validate(); err == nil {
		t.Error("expected validation error for empty method and nil handler")
	}
}

func TestRouteMiddlewareNames(t *testing.T) {
	r := Route{
		Middlewares: []Middleware{
			{Name: "auth", Handler: dummyHandler},
		},
	}
	names := r.MiddlewareNames()
	if len(names) != 1 || names[0] != "auth" {
		t.Errorf("expected [auth], got %v", names)
	}
}

func TestRouteHasAuth(t *testing.T) {
	tests := []struct {
		auth AuthRequirement
		want bool
	}{
		{AuthNone, false},
		{AuthOptional, false},
		{AuthRequired, true},
		{AuthDashboard, true},
		{AuthUnified, true},
		{AuthServiceKey, true},
		{AuthInternal, true},
	}

	for _, tt := range tests {
		r := Route{Auth: tt.auth}
		if got := r.HasAuth(); got != tt.want {
			t.Errorf("HasAuth() = %v, want %v", got, tt.want)
		}
	}
}

func TestRouteIsPublic(t *testing.T) {
	tests := []struct {
		public bool
		auth   AuthRequirement
		want   bool
	}{
		{true, AuthNone, true},
		{false, AuthNone, true},
		{false, AuthRequired, false},
	}

	for _, tt := range tests {
		r := Route{Public: tt.public, Auth: tt.auth}
		if got := r.IsPublic(); got != tt.want {
			t.Errorf("IsPublic() = %v, want %v", got, tt.want)
		}
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		parent   string
		child    string
		expected string
	}{
		{"", "/test", "/test"},
		{"/api", "/test", "/api/test"},
		{"/api/", "/test", "/api/test"},
		{"/api", "", "/api"},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := joinPath(tt.parent, tt.child)
		if result != tt.expected {
			t.Errorf("joinPath(%q, %q) = %q, want %q", tt.parent, tt.child, result, tt.expected)
		}
	}
}
