package messaging

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/templates"
	"github.com/pzsp-teams/lib/models"
)

type genericSender[Res any] struct {
	adapter senderAdapter[Res]
}

func newGenericSender[Res any](adapter senderAdapter[Res]) *genericSender[Res] {
	return &genericSender[Res]{
		adapter: adapter,
	}
}

func (s *genericSender[Res]) send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []Res {
	logger := initializers.Logger.With("dryRun", dryRun, "ignoreError", ignoreError, "total", len(messages))

	actions := s.planActions(ctx, messages)

	logger.Info("Starting bulk message send")

	var results []Res
	if dryRun {
		results = s.dryRunActions(actions)
	} else {
		results = s.executeActions(ctx, actions, ignoreError)
	}

	successCount := 0
	for _, r := range results {
		if s.adapter.getError(r) == nil {
			successCount++
		}
	}

	logger.Info("Bulk message send completed",
		"successful", successCount,
		"failed", len(messages)-successCount)

	return results
}

func (s *genericSender[Res]) planActions(ctx context.Context, messages map[string]string) []*action[Res] {
	actions := make([]*action[Res], 0, len(messages))

	for ref, content := range messages {
		processedContent, mentions, err := s.processMentions(ctx, ref, content)
		if err != nil {
			actions = append(actions, s.newErrorAction(ref, content, err))
			continue
		}

		data := &messageData{
			ref: ref,
			body: models.MessageBody{
				Content:     processedContent,
				ContentType: models.MessageContentTypeHTML,
				Mentions:    mentions,
			},
		}

		actions = append(actions, s.newSendAction(data))
	}

	return actions
}

func (s *genericSender[Res]) processMentions(ctx context.Context, ref, content string) (string, []models.Mention, error) {
	rawMentions := templates.ExtractMentions(content)
	if len(rawMentions) == 0 {
		return content, nil, nil
	}

	mentions, err := s.adapter.getMentions(ctx, ref, rawMentions)
	if err != nil {
		return "", nil, fmt.Errorf("%w for %s: %v", ErrMentionResolutionFailed, ref, err)
	}

	return templates.ReplaceMentions(content, mentions), mentions, nil
}

func (s *genericSender[Res]) executeActions(ctx context.Context, actions []*action[Res], ignoreError bool) []Res {
	logger := initializers.Logger
	results := make([]Res, 0, len(actions))
	skipRemaining := false

	for _, act := range actions {
		var result Res

		switch {
		case skipRemaining:
			logger.Debug("Skipping Message", s.getLogFields(act.data.ref)...)
			result = s.adapter.newErrorResult(act.data.ref, act.data.body.Content, ErrMessageSkipped)

		case act.run == nil:
			result = *act.result
			logFields := s.getLogFields(act.data.ref)
			logFields = append(logFields, "error", s.adapter.getError(result))
			logger.Error("Failed during planning", logFields...)
			if !ignoreError {
				skipRemaining = true
			}

		default:
			logger.Debug("Sending message", s.getLogFields(act.data.ref)...)
			result = act.run(ctx, act.data)

			if err := s.adapter.getError(result); err != nil {
				logFields := s.getLogFields(act.data.ref)
				logFields = append(logFields, "error", err)
				logger.Error("Failed to send message", logFields...)
				if !ignoreError {
					skipRemaining = true
				}
			} else {
				logger.Info("Message sent successfully", s.getLogFields(act.data.ref)...)
			}
		}

		results = append(results, result)
	}

	return results
}

func (s *genericSender[Res]) dryRunActions(actions []*action[Res]) []Res {
	results := make([]Res, 0, len(actions))
	for _, act := range actions {
		if act.run == nil {
			results = append(results, *act.result)
		} else {
			result := s.adapter.newSuccessResult(act.data.ref, act.data.body.Content)
			results = append(results, result)
		}
	}

	return results
}

func (s *genericSender[Res]) newSendAction(data *messageData) *action[Res] {
	return &action[Res]{
		data:   data,
		run:    s.sendMessage,
		result: nil,
	}
}

func (s *genericSender[Res]) newErrorAction(ref, content string, err error) *action[Res] {
	result := s.adapter.newErrorResult(ref, content, err)
	return &action[Res]{
		data: &messageData{
			ref: ref,
			body: models.MessageBody{
				Content: content,
			},
		},
		run:    nil,
		result: &result,
	}
}

func (s *genericSender[Res]) sendMessage(ctx context.Context, data *messageData) Res {
	msg, err := s.adapter.sendMessage(ctx, data.ref, data.body)
	if err != nil {
		return s.adapter.newErrorResult(data.ref, "", fmt.Errorf("%w to %s: %v", ErrMessageSendFailed, data.ref, err))
	}
	return s.adapter.newSuccessResult(data.ref, msg.Content)
}

func (s *genericSender[Res]) getLogFields(ref string) []any {
	fieldMap := s.adapter.getLogFields(ref)
	fields := make([]any, 0, len(fieldMap)*2)
	for k, v := range fieldMap {
		fields = append(fields, k, v)
	}
	return fields
}
