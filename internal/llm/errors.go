package llm

import "errors"

var (
	// ErrNoProvidersAvailable indicates no LLM providers could be initialized
	ErrNoProvidersAvailable = errors.New("no LLM providers available")

	// ErrAllProvidersFailed indicates all providers failed to generate a response
	ErrAllProvidersFailed = errors.New("all LLM providers failed")

	// ErrInvalidModel indicates the requested model is not available
	ErrInvalidModel = errors.New("invalid or unavailable model")

	// ErrTimeout indicates the generation request timed out
	ErrTimeout = errors.New("generation request timed out")

	// ErrInvalidResponse indicates the LLM returned an unparseable response
	ErrInvalidResponse = errors.New("invalid LLM response")
)
