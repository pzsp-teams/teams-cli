package creator

import "context"

// ExecuteActions runs all actions and collects results.
// For each action, it calls action.Run to execute the operation, handles nil results
// by calling nilResultFactory, logs the result, and collects all results.
func ExecuteActions[B any, R any](
	ctx context.Context,
	actions []Action[B, R],
	nilResultFactory func(action Action[B, R]) R,
	logFunc func(result *R),
) []R {
	results := make([]R, 0, len(actions))
	for _, act := range actions {
		result := act.Run(ctx, act.Body)
		if result == nil {
			r := nilResultFactory(act)
			result = &r
		}
		logFunc(result)
		results = append(results, *result)
	}
	return results
}

// DryRunActions returns pre-computed results without executing actions.
// This is used in dry-run mode to show what would happen without making actual changes.
// For each action, it uses the pre-computed action.Result, handles nil results
// by calling nilResultFactory, logs the result, and collects all results.
func DryRunActions[B any, R any](
	actions []Action[B, R],
	nilResultFactory func(action Action[B, R]) R,
	logFunc func(result *R),
) []R {
	results := make([]R, 0, len(actions))
	for _, act := range actions {
		res := act.Result
		if res == nil {
			r := nilResultFactory(act)
			res = &r
		}
		results = append(results, *res)
		logFunc(res)
	}
	return results
}

// PlanActions transforms bodies into actions using the provided planner function.
// The planner function is called for each body to determine what action should be taken.
func PlanActions[B any, R any](
	ctx context.Context,
	bodies []B,
	planner func(ctx context.Context, body *B) Action[B, R],
) []Action[B, R] {
	actions := make([]Action[B, R], 0, len(bodies))
	for i := range bodies {
		act := planner(ctx, &bodies[i])
		actions = append(actions, act)
	}
	return actions
}
