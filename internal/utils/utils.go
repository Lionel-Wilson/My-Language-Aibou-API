package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/models"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the structure of an error response.
// It contains a status code, a message, and an optional list of errors.
type ErrorResponse struct {
	StatusCode int      `json:"statusCode" example:"422"`
	Message    string   `json:"message" example:"Validation failed"`
	Errors     []string `json:"errors,omitempty"`
}

// TrimWhitespace trims leading and trailing whitespace from all string fields in a given struct.
func TrimWhitespace(v interface{}) {
	val := reflect.ValueOf(v).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String {
			field.SetString(strings.TrimSpace(field.String()))
		}
	}
}

// NewErrorResponse creates a new error response with the provided status code, message, and errors.
// It sends a JSON response with these details to the client.
func NewErrorResponse(c *gin.Context, statusCode int, message string, errors []string) {
	c.JSON(statusCode, ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
		Errors:     errors,
	})
}

func ServerErrorResponse(c *gin.Context, err error, msg string) {
	var message string
	if msg != "" {
		message = msg
	} else {
		message = "Something went wrong. Please try again later."
	}
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Errors:     []string{err.Error()},
	})
}

func ExtractIntegerCookie(c *gin.Context, cookieName string) (int, error) {
	cookieValueAsString, err := c.Request.Cookie(cookieName)
	if err != nil {
		return 0, err
	}

	cookieValueAsInt, err := strconv.Atoi(cookieValueAsString.Value)
	if err != nil {
		return 0, err
	}

	return cookieValueAsInt, nil
}

func MakeOpenAIApiRequest(body *strings.Reader, context *gin.Context, apiKey string) (string, error) {
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		ServerErrorResponse(context, err, "Failed to create request")
		return "", err
	}

	req.Header.Add("Content-Type", `application/json`)
	req.Header.Add("Authorization", `Bearer `+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ServerErrorResponse(context, err, "Failed to breakdown phrase")
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		ServerErrorResponse(context, err, "Failed to read AI response body")
		return "", err
	}
	var aiResponse models.ChatCompletion

	err = json.Unmarshal(responseBody, &aiResponse)
	if err != nil {
		ServerErrorResponse(context, err, "Failed to unmarshal json body")
		return "", err
	}

	return aiResponse.Choices[0].Message.Content, nil
}
