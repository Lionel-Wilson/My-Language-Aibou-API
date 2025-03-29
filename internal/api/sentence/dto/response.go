package dto

type ExplainSentenceResponse struct {
	ExpressionAnswer string `json:"expression_answer"`
	Error            string `json:"error"`
}

func ToExplainSentenceResponse(expressionAnswer, error string) ExplainSentenceResponse {
	return ExplainSentenceResponse{
		ExpressionAnswer: expressionAnswer,
		Error:            error,
	}
}
