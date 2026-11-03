package agent

import (
	"context"
	"errors"
	"fmt"
)

// AbortCause names why an agent run's context was cancelled. Every cancellation
// site should attach one so the activity trace can report the real initiator.
type AbortCause struct {
	Code  string
	Short string
	Text  string
}

func (c *AbortCause) Error() string { return c.Text }

func cause(code, short, text string) *AbortCause {
	return &AbortCause{Code: code, Short: short, Text: text}
}

// Causef builds a cause that includes runtime-specific detail.
func Causef(code, short, format string, args ...any) *AbortCause {
	return &AbortCause{Code: code, Short: short, Text: fmt.Sprintf(format, args...)}
}

var (
	// Task-level execution context.
	AbortPausedByUser = cause("paused_by_user", "User paused task",
		"The user paused the task via task control API (POST /api/tasks/{id}/control, action=pause). This Planner/Worker run was cancelled; running intents return to frontier (open) to be claimed and re-executed when resumed")
	AbortPausedByOrchestrator = cause("paused_by_orchestrator", "Orchestrator paused task",
		"The orchestrator agent called pause_task tool to pause this task. This Planner/Worker run was cancelled; running intents return to frontier (open) and re-execute upon resume")
	AbortTaskDeleted = cause("task_deleted", "Task deleted",
		"Task is being deleted (DELETE /api/tasks/{id}); the deletion barrier cancelled the running Planner, Worker, and Main Agent; this run output will not be used")
	AbortPausedOnReload = cause("paused_on_reload", "Task paused state restored on reload",
		"The backend restored the task pause state from database on startup. This run was cancelled; normally no agent runs during recovery")
	AbortGoalMet = cause("goal_met", "Planner determined task goal is met",
		"Planner determined task goal was met and set task to done, then cancelled running workers; these intents are marked stopped rather than failed")
	AbortSettleDrainTimeout = cause("settle_drain_timeout", "Task timeout wrap-up drain expired",
		"Task reached timeout and waited for running workers to gracefully wrap up, but the 90s drain grace period was exceeded, so hard cancellation was executed; intents are marked exhausted, facts and assets written during wrap-up are retained")

	// Per-work context.
	AbortKilledByPlanner = cause("killed_by_planner", "Planner terminated this intent",
		"Planner actively terminated this intent via kill_work, usually indicating an off-track direction or no further value; the intent is marked stopped and will not be reclaimed")
	AbortWorkPausedByUser = cause("work_paused_by_user", "User paused this worker intent",
		"The user paused the running worker. This invocation was cancelled and the intent changed to paused; all recorded intents, facts, findings, and activities are retained, and re-executed from start upon resume")
	AbortWorkCancelledByUser = cause("work_cancelled_by_user", "User deleted this worker intent",
		"The user deleted the running worker. This invocation was cancelled; after worker exits write lock, the server processes the intent according to the selected delete mode: soft delete marks as deleted and retains outputs, hard delete cascades and removes the intent and downstream nodes")
	AbortWorkFinished = cause("work_finished", "Worker finished normally",
		"Worker finished normally, and engine released context resources in detachWork. This is not an interruption; if it appears in an interruption message, cancellation and teardown events raced")
	AbortPausedRaceGuard = cause("paused_race_guard", "Task paused; refused to start run",
		"Task is in paused state; engine refused to issue new execution context to prevent race between claim and pause; claimed intents return to frontier")

	// Main Agent and standalone conversation contexts.
	AbortChatStoppedByUser = cause("chat_stopped_by_user", "User stopped this chat turn",
		"The user clicked stop to abort this Main Agent or Chat Agent run. Generated activities are retained, and next message can be sent")
	AbortChatPausedWithTask = cause("chat_paused_with_task", "Task paused; aborted chat turn",
		"When the user paused the task, the running Main Agent chat was also cancelled. Generated activities are retained; messages are not automatically replayed upon resume")
	AbortChatTurnFinished = cause("chat_turn_finished", "Chat turn finished normally",
		"Chat turn finished normally, and server is releasing turn context resources. This is not an interruption; if it appears in an interruption message, cancellation and teardown events raced")

	// Process-level and per-run hard backstop.
	AbortShutdown = cause("shutdown", "Backend process shutting down",
		"Backend process received SIGINT or SIGTERM and is restarting, updating, or shutting down. All running agents will be cancelled; running intents will reset to open upon restart and re-execute")
	AbortRunHardTimeout = cause("run_hard_timeout", "Single run hard timeout triggered",
		"Single run exceeded soft wall-clock budget and extra grace period, indicating a model request or tool did not return for a long time, preventing normal turn boundary wrap-up. Please check the last tool call before interruption")
)

// AbortReason resolves the named cause attached to a cancelled run context.
func AbortReason(ctx context.Context) (code, short, text string, ok bool) {
	c := context.Cause(ctx)
	if c == nil {
		return "", "", "", false
	}
	var ac *AbortCause
	if errors.As(c, &ac) {
		return ac.Code, ac.Short, ac.Text, true
	}
	switch {
	case errors.Is(c, context.DeadlineExceeded):
		return "deadline_exceeded", "Upstream context deadline exceeded",
			"Upstream context reached deadline, but caller did not attach a named cause via WithTimeoutCause: " + c.Error(), true
	case errors.Is(c, context.Canceled):
		return "canceled_no_cause", "Caller attached no named cause",
			"Upstream context was cancelled, but caller did not attach a named cause via context.WithCancelCause; please register cause in agent/cancelcause.go and wire to this cancellation point", true
	default:
		return "other", firstLine(c.Error(), 80), c.Error(), true
	}
}
