package creator

import "context"

// Action represents a planned operation with deferred execution
// B is the body type containing the request data
// R is the result type representing the outcome
type Action[B any, R any] struct {
	Body   B
	Run    func(ctx context.Context, body B) *R
	Result *R
}

// StaticAction creates an action that always returns a pre-computed result
// This is used for operations that don't need actual execution, such as:
// - Resources that already exist
// - Validation errors detected during planning
// - Operations blocked by previous failures
func StaticAction[B any, R any](body B, result R) Action[B, R] {
	return Action[B, R]{
		Body:   body,
		Result: &result,
		Run: func(ctx context.Context, b B) *R {
			return &result
		},
	}
}
