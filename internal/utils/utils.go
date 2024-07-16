package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

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

	var errorDetails []string
	if err != nil {
		errorDetails = append(errorDetails, err.Error())
	} else {
		errorDetails = append(errorDetails, "Unknown error")
	}
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Errors:     errorDetails,
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

func MakeOpenAIApiRequest(body *strings.Reader, context *gin.Context, apiKey string) (models.ChatCompletion, error) {
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		fmt.Println("Failed to create request")
		return models.ChatCompletion{}, err
	}

	req.Header.Add("Content-Type", `application/json`)
	req.Header.Add("Authorization", `Bearer `+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to breakdown phrase")
		return models.ChatCompletion{}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read AI response body:")
		fmt.Println(string(responseBody))
		return models.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		fmt.Println("OpenAI error response body:")
		fmt.Println(string(responseBody))
		return models.ChatCompletion{}, err
	}

	var aiResponse models.ChatCompletion

	err = json.Unmarshal(responseBody, &aiResponse)
	if err != nil {
		fmt.Println("Failed to unmarshal json body")
		return models.ChatCompletion{}, err
	}

	if len(aiResponse.Choices) == 0 {
		fmt.Println("OpenAI API response contains no choices")
		err = fmt.Errorf("OpenAI API response contains no choices")
		return models.ChatCompletion{}, err
	}

	return aiResponse, nil
}

// containsNumber checks if a given string contains a number.
func ContainsNumber(s string) bool {
	for _, ch := range s {
		if unicode.IsDigit(ch) {
			return true
		}
	}
	return false
}
