package sender

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/initializers"
	"github.com/pzsp-teams/teams-cli/internal/templates"
)

// GenericSender handles message sending with mention resolution and error handling
type GenericSender[Res Result] struct {
	adapter    Adapter
	newResult  func(ref, message string, err error) Res
	senderType Type
}

// NewGenericSender creates a new GenericSender with the provided adapter and result factory
func NewGenericSender[Res Result](
	adapter Adapter,
	newResult func(ref, message string, err error) Res,
	st Type,
) *GenericSender[Res] {
	return &GenericSender[Res]{
		adapter:    adapter,
		newResult:  newResult,
		senderType: st,
	}
}

// Send executes bulk message sending with the provided options
func (s *GenericSender[Res]) Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []Res {
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
		if r.GetError() == nil {
			successCount++
		}
	}

	logger.Info("Bulk message send completed",
		"successful", successCount,
		"failed", len(messages)-successCount)

	return results
}

func (s *GenericSender[Res]) planActions(ctx context.Context, messages map[string]string) []*Action[Res] {
	actions := make([]*Action[Res], 0, len(messages))

	for ref, content := range messages {
		processedContent, mentions, err := s.processMentions(ctx, ref, content)
		if err != nil {
			actions = append(actions, s.newErrorAction(ref, content, err))
			continue
		}

		data := &MessageData{
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

func (s *GenericSender[Res]) processMentions(ctx context.Context, ref, content string) (string, []models.Mention, error) {
	rawMentions := templates.ExtractMentions(content)
	if len(rawMentions) == 0 {
		return content, nil, nil
	}

	mentions, err := s.adapter.GetMentions(ctx, ref, rawMentions)
	if err != nil {
		return "", nil, fmt.Errorf("%w for %s: %v", ErrMentionResolutionFailed, ref, err)
	}

	return templates.ReplaceMentions(content, mentions), mentions, nil
}

func (s *GenericSender[Res]) executeActions(ctx context.Context, actions []*Action[Res], ignoreError bool) []Res {
	logger := initializers.Logger
	results := make([]Res, 0, len(actions))
	skipRemaining := false

	for _, act := range actions {
		var result Res

		switch {
		case skipRemaining:
			logger.Debug("Skipping Message", s.senderType.String(), act.data.ref)
			result = s.newResult(act.data.ref, act.data.body.Content, ErrMessageSkipped)

		case act.run == nil:
			result = *act.result
			logger.Error("Failed during planning", s.senderType.String(), act.data.ref, "error", result.GetError())
			if !ignoreError {
				skipRemaining = true
			}

		default:
			logger.Debug("Sending message", s.senderType.String(), act.data.ref)
			result = act.run(ctx, act.data)

			if err := result.GetError(); err != nil {
				logger.Error("Failed to send message", s.senderType.String(), act.data.ref, "error", err)
				if !ignoreError {
					skipRemaining = true
				}
			} else {
				logger.Info("Message sent successfully", s.senderType.String(), act.data.ref)
			}
		}

		results = append(results, result)
	}

	return results
}

func (s *GenericSender[Res]) dryRunActions(actions []*Action[Res]) []Res {
	logger := initializers.Logger
	results := make([]Res, 0, len(actions))
	for _, act := range actions {
		if act.run == nil {
			result := *act.result
			results = append(results, result)
			logger.Debug("Dry run: failed during planning", s.senderType.String(), act.data.ref, "error", result.GetError())
		} else {
			logger.Debug("Dry run: would send message", s.senderType.String(), act.data.ref)
			result := s.newResult(act.data.ref, act.data.body.Content, nil)
			results = append(results, result)
		}
	}

	return results
}

func (s *GenericSender[Res]) newSendAction(data *MessageData) *Action[Res] {
	return &Action[Res]{
		data:   data,
		run:    s.sendMessage,
		result: nil,
	}
}

func (s *GenericSender[Res]) newErrorAction(ref, content string, err error) *Action[Res] {
	result := s.newResult(ref, content, err)
	return &Action[Res]{
		data: &MessageData{
			ref: ref,
			body: models.MessageBody{
				Content: content,
			},
		},
		run:    nil,
		result: &result,
	}
}

func (s *GenericSender[Res]) sendMessage(ctx context.Context, data *MessageData) Res {
	msg, err := s.adapter.SendMessage(ctx, data.ref, data.body)
	if err != nil {
		return s.newResult(data.ref, "", fmt.Errorf("%w to %s: %v", ErrMessageSendFailed, data.ref, err))
	}
	return s.newResult(data.ref, msg.Content, nil)
}
