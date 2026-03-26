package avoresponse

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// Response represents a successful API response with generic data type.
// Meta and Links are optional and omitted from JSON when nil.
type Response[T any] struct {
	Data  T      `json:"data"`
	Meta  *Meta  `json:"meta,omitempty"`
	Links *Links `json:"links,omitempty"`
}

// ErrorResponse represents an API error response containing one or more errors.
type ErrorResponse struct {
	Errors []ErrorItem `json:"errors"`
}

// ErrorItem represents a single error entry with a machine-readable code
// and a human-readable message.
type ErrorItem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta holds pagination metadata for list responses.
type Meta struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"per_page"`
	TotalCount  int  `json:"total_count"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// Links holds pagination links for list responses.
// Next and Previous are omitted when empty (first/last page).
type Links struct {
	Self     string `json:"self"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// PaginationParams holds the parameters needed to build pagination meta and links.
type PaginationParams struct {
	Page       int
	PerPage    int
	TotalCount int
	BaseURL    string
}

// OK sends a 200 response with the given data.
func OK[T any](c fiber.Ctx, data T) error {
	return c.Status(fiber.StatusOK).JSON(Response[T]{Data: data})
}

// OKWithMeta sends a 200 response with data and meta information.
func OKWithMeta[T any](c fiber.Ctx, data T, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(Response[T]{
		Data: data,
		Meta: meta,
	})
}

// Created sends a 201 response with the given data.
func Created[T any](c fiber.Ctx, data T) error {
	return c.Status(fiber.StatusCreated).JSON(Response[T]{Data: data})
}

// NoContent sends a 204 response with no body.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Fail sends an error response with the given HTTP status and error items.
func Fail(c fiber.Ctx, status int, items ...ErrorItem) error {
	return c.Status(status).JSON(ErrorResponse{Errors: items})
}

// OKWithPagination sends a 200 response with data, pagination meta, and links.
func OKWithPagination[T any](c fiber.Ctx, data T, params PaginationParams) error {
	meta := BuildMeta(params.Page, params.PerPage, params.TotalCount)
	links := BuildLinks(params.BaseURL, params.Page, params.PerPage, params.TotalCount)

	return c.Status(fiber.StatusOK).JSON(Response[T]{
		Data:  data,
		Meta:  meta,
		Links: links,
	})
}

// BuildMeta creates a Meta struct from pagination parameters.
// HasNext is true when page*perPage < totalCount (more items exist).
// HasPrevious is true when page > 1.
func BuildMeta(page, perPage, totalCount int) *Meta {
	return &Meta{
		Page:        page,
		PerPage:     perPage,
		TotalCount:  totalCount,
		HasNext:     page*perPage < totalCount,
		HasPrevious: page > 1,
	}
}

// BuildLinks creates pagination Links from the base URL and parameters.
// Self always points to the current page.
// Next is set only if there are more pages.
// Previous is set only if the current page is not the first.
func BuildLinks(baseURL string, page, perPage, totalCount int) *Links {
	links := &Links{
		Self: fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page, perPage),
	}

	if page*perPage < totalCount {
		links.Next = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page+1, perPage)
	}

	if page > 1 {
		links.Previous = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page-1, perPage)
	}

	return links
}
