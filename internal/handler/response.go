package handler

import (
	"github.com/labstack/echo/v4"
)

type apiResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type paginatedData struct {
	Data  any `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

func success(c echo.Context, status int, data any) error {
	return c.JSON(status, apiResponse{Success: true, Data: data})
}

func successPaginated(c echo.Context, status int, data any, total, page, limit int) error {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return c.JSON(status, apiResponse{
		Success: true,
		Data: paginatedData{
			Data:  data,
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: pages,
		},
	})
}

func fail(c echo.Context, status int, message string) error {
	return c.JSON(status, apiResponse{
		Success: false,
		Error:   &apiError{Code: "error", Message: message},
	})
}
