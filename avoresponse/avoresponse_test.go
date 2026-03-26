package avoresponse_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/bilustek/avokado/avoerror"
	"github.com/bilustek/avokado/avoresponse"
	"github.com/gofiber/fiber/v3"
)

func setupApp() *fiber.App {
	return fiber.New()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

func TestResponse_JSONMarshal_HasDataMetaLinks(t *testing.T) {
	t.Parallel()

	meta := &avoresponse.Meta{
		Page:        1,
		PerPage:     10,
		TotalCount:  100,
		HasNext:     true,
		HasPrevious: false,
	}

	links := &avoresponse.Links{
		Self: "/api/items?page=1&per_page=10",
		Next: "/api/items?page=2&per_page=10",
	}

	resp := avoresponse.Response[[]string]{
		Data:  []string{"a", "b"},
		Meta:  meta,
		Links: links,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := raw["data"]; !ok {
		t.Error("expected 'data' key in JSON output")
	}

	if _, ok := raw["meta"]; !ok {
		t.Error("expected 'meta' key in JSON output")
	}

	if _, ok := raw["links"]; !ok {
		t.Error("expected 'links' key in JSON output")
	}
}

func TestResponse_JSONMarshal_OmitsEmptyMetaAndLinks(t *testing.T) {
	t.Parallel()

	resp := avoresponse.Response[string]{
		Data: "hello",
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := raw["meta"]; ok {
		t.Error("expected 'meta' to be omitted when nil")
	}

	if _, ok := raw["links"]; ok {
		t.Error("expected 'links' to be omitted when nil")
	}
}

func TestOK_Returns200WithData(t *testing.T) {
	t.Parallel()

	app := setupApp()

	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	app.Get("/test", func(c fiber.Ctx) error {
		return avoresponse.OK(c, item{ID: 1, Name: "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result map[string]json.RawMessage

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := result["data"]; !ok {
		t.Error("expected 'data' key in response")
	}

	var data item

	if err := json.Unmarshal(result["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if data.ID != 1 || data.Name != "test" {
		t.Errorf("unexpected data: %+v", data)
	}
}

func TestOKWithMeta_Returns200WithDataAndMeta(t *testing.T) {
	t.Parallel()

	app := setupApp()

	app.Get("/test", func(c fiber.Ctx) error {
		meta := &avoresponse.Meta{
			Page:       1,
			PerPage:    10,
			TotalCount: 50,
			HasNext:    true,
		}

		return avoresponse.OKWithMeta(c, []string{"a", "b"}, meta)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result map[string]json.RawMessage

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := result["data"]; !ok {
		t.Error("expected 'data' key")
	}

	if _, ok := result["meta"]; !ok {
		t.Error("expected 'meta' key")
	}
}

func TestCreated_Returns201WithData(t *testing.T) {
	t.Parallel()

	app := setupApp()

	app.Post("/test", func(c fiber.Ctx) error {
		return avoresponse.Created(c, map[string]string{"id": "123"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result map[string]json.RawMessage

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := result["data"]; !ok {
		t.Error("expected 'data' key")
	}
}

func TestNoContent_Returns204WithNoBody(t *testing.T) {
	t.Parallel()

	app := setupApp()

	app.Delete("/test", func(c fiber.Ctx) error {
		return avoresponse.NoContent(c)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) != 0 {
		t.Errorf("expected empty body for 204, got %q", string(body))
	}
}

func TestFail_ReturnsErrorResponse(t *testing.T) {
	t.Parallel()

	app := setupApp()

	app.Get("/test", func(c fiber.Ctx) error {
		return avoresponse.Fail(c, 400,
			avoresponse.ErrorItem{Code: "validation-error", Message: "email is required"},
			avoresponse.ErrorItem{Code: "validation-error", Message: "name too short"},
		)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result avoresponse.ErrorResponse

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result.Errors))
	}

	if result.Errors[0].Code != "validation-error" {
		t.Errorf("expected code 'validation-error', got %q", result.Errors[0].Code)
	}

	if result.Errors[0].Message != "email is required" {
		t.Errorf("expected message 'email is required', got %q", result.Errors[0].Message)
	}
}

func TestErrorResponse_JSONMarshal(t *testing.T) {
	t.Parallel()

	resp := avoresponse.ErrorResponse{
		Errors: []avoresponse.ErrorItem{
			{Code: "not-found", Message: "resource not found"},
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	errorsRaw, ok := raw["errors"]
	if !ok {
		t.Fatal("expected 'errors' key in JSON")
	}

	var items []avoresponse.ErrorItem

	if err := json.Unmarshal(errorsRaw, &items); err != nil {
		t.Fatalf("unmarshal errors: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 error item, got %d", len(items))
	}

	if items[0].Code != "not-found" {
		t.Errorf("expected code 'not-found', got %q", items[0].Code)
	}

	if items[0].Message != "resource not found" {
		t.Errorf("expected message 'resource not found', got %q", items[0].Message)
	}
}

func TestMeta_HasExpectedFields(t *testing.T) {
	t.Parallel()

	meta := avoresponse.Meta{
		Page:        2,
		PerPage:     25,
		TotalCount:  100,
		HasNext:     true,
		HasPrevious: true,
	}

	if meta.Page != 2 {
		t.Errorf("expected Page 2, got %d", meta.Page)
	}

	if meta.PerPage != 25 {
		t.Errorf("expected PerPage 25, got %d", meta.PerPage)
	}

	if meta.TotalCount != 100 {
		t.Errorf("expected TotalCount 100, got %d", meta.TotalCount)
	}

	if !meta.HasNext {
		t.Error("expected HasNext true")
	}

	if !meta.HasPrevious {
		t.Error("expected HasPrevious true")
	}
}

func TestMeta_JSONTags(t *testing.T) {
	t.Parallel()

	meta := avoresponse.Meta{
		Page:        1,
		PerPage:     10,
		TotalCount:  50,
		HasNext:     true,
		HasPrevious: false,
	}

	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedKeys := []string{"page", "per_page", "total_count", "has_next", "has_previous"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

func TestLinks_HasExpectedFields(t *testing.T) {
	t.Parallel()

	links := avoresponse.Links{
		Self:     "/api/items?page=1&per_page=10",
		Next:     "/api/items?page=2&per_page=10",
		Previous: "/api/items?page=0&per_page=10",
	}

	if links.Self != "/api/items?page=1&per_page=10" {
		t.Errorf("unexpected Self: %q", links.Self)
	}

	if links.Next != "/api/items?page=2&per_page=10" {
		t.Errorf("unexpected Next: %q", links.Next)
	}

	if links.Previous != "/api/items?page=0&per_page=10" {
		t.Errorf("unexpected Previous: %q", links.Previous)
	}
}

func TestLinks_OmitsEmptyNextPrevious(t *testing.T) {
	t.Parallel()

	links := avoresponse.Links{
		Self: "/api/items?page=1&per_page=10",
	}

	b, err := json.Marshal(links)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := raw["next"]; ok {
		t.Error("expected 'next' to be omitted when empty")
	}

	if _, ok := raw["previous"]; ok {
		t.Error("expected 'previous' to be omitted when empty")
	}
}

func TestBuildMeta_HasNextTrue(t *testing.T) {
	t.Parallel()

	// page=1, perPage=10, total=25 -> 1*10=10 < 25 -> HasNext=true
	meta := avoresponse.BuildMeta(1, 10, 25)

	if !meta.HasNext {
		t.Error("expected HasNext=true when page*perPage < totalCount")
	}

	if meta.HasPrevious {
		t.Error("expected HasPrevious=false when page=1")
	}
}

func TestBuildMeta_HasNextFalse(t *testing.T) {
	t.Parallel()

	// page=3, perPage=10, total=25 -> 3*10=30 >= 25 -> HasNext=false
	meta := avoresponse.BuildMeta(3, 10, 25)

	if meta.HasNext {
		t.Error("expected HasNext=false when page*perPage >= totalCount")
	}

	if !meta.HasPrevious {
		t.Error("expected HasPrevious=true when page > 1")
	}
}

func TestBuildMeta_HasPreviousTrue(t *testing.T) {
	t.Parallel()

	meta := avoresponse.BuildMeta(2, 10, 50)

	if !meta.HasPrevious {
		t.Error("expected HasPrevious=true when page > 1")
	}
}

func TestBuildMeta_HasPreviousFalse(t *testing.T) {
	t.Parallel()

	meta := avoresponse.BuildMeta(1, 10, 50)

	if meta.HasPrevious {
		t.Error("expected HasPrevious=false when page == 1")
	}
}

func TestBuildLinks_FirstPage_NoPreviousLink(t *testing.T) {
	t.Parallel()

	links := avoresponse.BuildLinks("/api/items", 1, 10, 25)

	if links.Self == "" {
		t.Error("expected Self link to be set")
	}

	if links.Previous != "" {
		t.Error("expected Previous to be empty on first page")
	}

	if links.Next == "" {
		t.Error("expected Next link on first page when more pages exist")
	}
}

func TestBuildLinks_LastPage_NoNextLink(t *testing.T) {
	t.Parallel()

	// page=3, perPage=10, total=25 -> last page
	links := avoresponse.BuildLinks("/api/items", 3, 10, 25)

	if links.Next != "" {
		t.Error("expected Next to be empty on last page")
	}

	if links.Previous == "" {
		t.Error("expected Previous link on last page")
	}
}

func TestBuildLinks_MiddlePage_BothLinks(t *testing.T) {
	t.Parallel()

	links := avoresponse.BuildLinks("/api/items", 2, 10, 50)

	if links.Next == "" {
		t.Error("expected Next link on middle page")
	}

	if links.Previous == "" {
		t.Error("expected Previous link on middle page")
	}

	if links.Self == "" {
		t.Error("expected Self link")
	}
}

func TestBuildLinks_LinkFormat(t *testing.T) {
	t.Parallel()

	links := avoresponse.BuildLinks("/api/items", 2, 10, 50)

	expected := "/api/items?page=2&per_page=10"
	if links.Self != expected {
		t.Errorf("expected Self=%q, got %q", expected, links.Self)
	}

	expectedNext := "/api/items?page=3&per_page=10"
	if links.Next != expectedNext {
		t.Errorf("expected Next=%q, got %q", expectedNext, links.Next)
	}

	expectedPrev := "/api/items?page=1&per_page=10"
	if links.Previous != expectedPrev {
		t.Errorf("expected Previous=%q, got %q", expectedPrev, links.Previous)
	}
}

func TestOKWithPagination_MetaAndLinksCorrect(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	app.Get("/items", func(c fiber.Ctx) error {
		params := avoresponse.PaginationParams{
			Page:       2,
			PerPage:    10,
			TotalCount: 50,
			BaseURL:    "/api/items",
		}

		return avoresponse.OKWithPagination(c, []string{"item1", "item2"}, params)
	})

	req := httptest.NewRequest("GET", "/items", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data  []string           `json:"data"`
		Meta  *avoresponse.Meta  `json:"meta"`
		Links *avoresponse.Links `json:"links"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if result.Meta == nil {
		t.Fatal("expected meta to be present")
	}

	if result.Meta.Page != 2 {
		t.Errorf("expected meta.page=2, got %d", result.Meta.Page)
	}

	if result.Meta.PerPage != 10 {
		t.Errorf("expected meta.per_page=10, got %d", result.Meta.PerPage)
	}

	if result.Meta.TotalCount != 50 {
		t.Errorf("expected meta.total_count=50, got %d", result.Meta.TotalCount)
	}

	if !result.Meta.HasNext {
		t.Error("expected HasNext=true (2*10=20 < 50)")
	}

	if !result.Meta.HasPrevious {
		t.Error("expected HasPrevious=true (page > 1)")
	}

	if result.Links == nil {
		t.Fatal("expected links to be present")
	}

	if result.Links.Self != "/api/items?page=2&per_page=10" {
		t.Errorf("unexpected Self link: %q", result.Links.Self)
	}

	if result.Links.Next != "/api/items?page=3&per_page=10" {
		t.Errorf("unexpected Next link: %q", result.Links.Next)
	}

	if result.Links.Previous != "/api/items?page=1&per_page=10" {
		t.Errorf("unexpected Previous link: %q", result.Links.Previous)
	}
}

func TestErrorHandler_FiberError(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result avoresponse.ErrorResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Code != "http-error" {
		t.Errorf("expected code 'http-error', got %q", result.Errors[0].Code)
	}

	if result.Errors[0].Message != "not found" {
		t.Errorf("expected message 'not found', got %q", result.Errors[0].Message)
	}
}

func TestErrorHandler_AvokadoError(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return avoerror.New("user not found").
			WithStatus(fiber.StatusNotFound).
			WithCode(avoerror.CodeNotFound)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result avoresponse.ErrorResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Code != string(avoerror.CodeNotFound) {
		t.Errorf("expected code %q, got %q", avoerror.CodeNotFound, result.Errors[0].Code)
	}

	if result.Errors[0].Message != "user not found" {
		t.Errorf("expected message 'user not found', got %q", result.Errors[0].Message)
	}
}

func TestErrorHandler_AvokadoError_DefaultStatus500(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return avoerror.New("something broke").
			WithCode(avoerror.CodeInternalError)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected status %d (default), got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestErrorHandler_UnknownError(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return errors.New("secret database password exposed") //nolint:err113
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result avoresponse.ErrorResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Code != "internal-error" {
		t.Errorf("expected code 'internal-error', got %q", result.Errors[0].Code)
	}

	// Real error message must NOT be exposed
	if result.Errors[0].Message != "internal server error" {
		t.Errorf("expected 'internal server error', got %q", result.Errors[0].Message)
	}

	// Verify the real message is NOT in the response body
	bodyStr := string(body)
	if contains(bodyStr, "secret database password exposed") {
		t.Error("real error message should NOT appear in response body")
	}
}

func TestErrorHandler_NilError(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{}),
	})

	// Fiber may call error handler with nil in edge cases
	app.Get("/test", func(c fiber.Ctx) error {
		// Manually invoke the error handler with nil
		return app.Config().ErrorHandler(c, nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestErrorHandler_WithLogger_ServerError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{
			Logger: logger,
		}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return errors.New("some error") //nolint:err113
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	logOutput := buf.String()
	if !contains(logOutput, "server error") {
		t.Error("expected 'server error' in log output")
	}
	if !contains(logOutput, "ERROR") {
		t.Error("expected ERROR level in log output")
	}
}

func TestErrorHandler_WithLogger_ClientError_LogDisabled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{
			Logger: logger,
		}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}

	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("expected no log output when LogClientErrors is false, got %q", logOutput)
	}
}

func TestErrorHandler_WithLogger_ClientError_LogEnabled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	app := fiber.New(fiber.Config{
		ErrorHandler: avoresponse.NewErrorHandler(&avoresponse.ErrorHTTPHandlerArgs{
			Logger:          logger,
			LogClientErrors: true,
		}),
	})

	app.Get("/test", func(_ fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}

	logOutput := buf.String()
	if !contains(logOutput, "client error") {
		t.Error("expected 'client error' in log output")
	}
	if !contains(logOutput, "WARN") {
		t.Error("expected WARN level in log output")
	}
}
