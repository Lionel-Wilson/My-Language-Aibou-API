package sentence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"

	"go.uber.org/zap"

	airesponses "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/airesponses/storage"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
)

var ErrOpenAiNoChoices = errors.New("OpenAI API response contains no choices")

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetSentenceExplanation(ctx context.Context, sentence string, nativeLanguage string, isDetailed bool) (*string, error)
	GetSentenceCorrection(ctx context.Context, sentence string, nativeLanguage string) (*string, error)
	ValidateSentence(sentence string) error
}

type service struct {
	logger       *zap.Logger
	openAiClient openai.Client
	responses    airesponses.Repository
}

func NewSentenceService(
	logger *zap.Logger,
	openAiClient openai.Client,
	responses airesponses.Repository,
) Service {
	return &service{
		logger:       logger,
		openAiClient: openAiClient,
		responses:    responses,
	}
}

// Bump the matching constant when the prompt for that kind changes.
const (
	promptVersionExplanation = 1
	promptVersionTranslation = 1
	promptVersionCorrection  = 1
)

func (s *service) getCachedResponse(ctx context.Context, kind airesponses.Kind, input, nativeLanguage string, promptVersion int) (string, bool) {
	return airesponses.GetCached(ctx, s.logger, s.responses, kind, input, nativeLanguage, promptVersion)
}

func (s *service) storeCachedResponse(ctx context.Context, kind airesponses.Kind, input, nativeLanguage string, promptVersion int, response string) {
	airesponses.StoreCached(ctx, s.logger, s.responses, kind, input, nativeLanguage, promptVersion, response)
}

func (s *service) GetSentenceCorrection(ctx context.Context, sentence string, nativeLanguage string) (*string, error) {
	if cached, ok := s.getCachedResponse(ctx, airesponses.KindCorrection, sentence, nativeLanguage, promptVersionCorrection); ok {
		return &cached, nil
	}

	jsonBody, err := s.sentenceToOpenAiSentenceCorrectionRequestBody(sentence, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	resp, responseBody, err := s.openAiClient.MakeRequest(ctx, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make open ai request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAI API returned non-OK status(%v): %s", resp.StatusCode, responseBody)
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhsal json body: %w", err)
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, ErrOpenAiNoChoices
	}

	s.logger.Info("Successfully got sentence correction",
		zap.String("sentence", sentence),
		zap.String("nativeLanguage", nativeLanguage),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	result := &OpenAIApiResponse.Choices[0].Message.Content
	s.storeCachedResponse(ctx, airesponses.KindCorrection, sentence, nativeLanguage, promptVersionCorrection, *result)

	return result, nil
}

func (s *service) GetSentenceExplanation(ctx context.Context, sentence string, nativeLanguage string, isDetailed bool) (*string, error) {
	kind := airesponses.KindTranslation
	promptVersion := promptVersionTranslation
	if isDetailed {
		kind = airesponses.KindExplanation
		promptVersion = promptVersionExplanation
	}

	if cached, ok := s.getCachedResponse(ctx, kind, sentence, nativeLanguage, promptVersion); ok {
		return &cached, nil
	}

	var jsonBody io.Reader
	var err error
	if isDetailed {
		jsonBody, err = s.sentenceToOpenAiExplanationRequestBody(sentence, nativeLanguage)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal openai request: %w", err)
		}
	} else {
		jsonBody, err = s.sentenceToOpenAiSimpleTranslationRequestBody(sentence, nativeLanguage)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal openai request: %w", err)
		}
	}

	resp, responseBody, err := s.openAiClient.MakeRequest(ctx, jsonBody)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned non-OK statusCode=%v responseBody=%s", resp.StatusCode, responseBody)
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhsal json body: %w", err)
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, ErrOpenAiNoChoices
	}

	s.logger.Info("Successfully got sentence explanation",
		zap.String("sentence", sentence),
		zap.String("nativeLanguage", nativeLanguage),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	result := &OpenAIApiResponse.Choices[0].Message.Content
	s.storeCachedResponse(ctx, kind, sentence, nativeLanguage, promptVersion, *result)

	return result, nil
}

//todo:return actual errors and convert to user friendlt message in handler
func (s *service) ValidateSentence(sentence string) error {
	if sentence == "" {
		return errors.New("Please provide a sentence")
	}

	if utf8.RuneCountInString(sentence) > 130 {
		return errors.New("The sentence must be less than 130 characters.")
	}

	return nil
}

func (s *service) sentenceToOpenAiSimpleTranslationRequestBody(sentence, userNativeLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Translate the following sentence into %s. Do not provide any explanation or context—only the translated sentence.\n\nSentence: %s",
		userNativeLanguage, sentence,
	)

	req := openai.OpenAIRequest{
		Model:       "gpt-4o",
		Temperature: 0.2,
		MaxTokens:   300,
		Messages: []openai.Message{
			{Role: "system", Content: "You are a precise translator. Only return the translation without explanations."},
			{Role: "user", Content: content},
		},
	}

	return request.JsonReader(&req)
}

const (
	plainLanguageRule = "Use plain, everyday language that a non-linguist can follow. Avoid technical grammar terms (such as imperative, subjunctive, indicative or direct object pronoun) and similar jargon. If a technical term is genuinely useful, immediately follow it with a simple, easy-to-understand synonym in brackets."

	responseLanguageSystemRule = "Always write your entire response, including all labels, headings and introductory phrases, in the language the user asks you to respond in. Only quoted words or sentences from the sentence being discussed may stay in their original language."
)

func nativeLanguageRule(userNativeLanguage string) string {
	return fmt.Sprintf(
		"The user only understands %[1]s. Write your ENTIRE response in %[1]s, including every introductory phrase, label and heading, and any grammar terms. Do not use any language other than %[1]s, except when quoting the original sentence, a corrected sentence, or words taken from them.",
		userNativeLanguage,
	)
}

func (s *service) sentenceToOpenAiExplanationRequestBody(sentence, userNativeLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Explain the meaning and how this sentence works.\n\n%s\n\n%s\n\nSentence: %s",
		nativeLanguageRule(userNativeLanguage), plainLanguageRule, sentence,
	)

	req := openai.OpenAIRequest{
		Model:       "gpt-4o",
		Temperature: 0.4,
		MaxTokens:   800,
		Messages: []openai.Message{
			{Role: "system", Content: "You are a friendly language tutor explaining sentences to everyday learners. " + responseLanguageSystemRule},
			{Role: "user", Content: content},
		},
	}

	return request.JsonReader(&req)
}

func (s *service) sentenceToOpenAiSentenceCorrectionRequestBody(sentence, userNativeLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Is this sentence correct? If not, correct it and briefly explain why. First, state what you assume the user is trying to say, with a short introductory phrase meaning \"Assuming you're trying to say\". Do not ask follow-up questions or encourage further conversation. Just provide the assumed meaning, correction, and explanation in a single, complete answer, as if you're a language teacher teaching a native %s speaker.\n\n%s\n\n%s\n\nSentence: %s",
		userNativeLanguage, nativeLanguageRule(userNativeLanguage), plainLanguageRule, sentence,
	)

	req := openai.OpenAIRequest{
		Model:       "gpt-4o",
		Temperature: 0.4,
		MaxTokens:   800,
		Messages: []openai.Message{
			{Role: "system", Content: "You are a concise language assistant that explains sentence corrections in a clear and brief manner without engaging in back-and-forth conversation. " + responseLanguageSystemRule},
			{Role: "user", Content: content},
		},
	}

	return request.JsonReader(&req)
}
