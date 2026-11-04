package control

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var reAgentKey = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidateID ensures a string is a positive decimal integer in signed 64-bit range.
func ValidateID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("ID cannot be empty")
	}
	// Strict decimal digit check: no signs, no exponents, no decimals
	for _, ch := range id {
		if ch < '0' || ch > '9' {
			return errors.New("must be a positive decimal string")
		}
	}
	val, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return errors.New("exceeds 64-bit range")
		}
		return errors.New("must be a positive decimal string")
	}
	if val <= 0 {
		return errors.New("must be positive")
	}
	return nil
}

// ValidateAgentKey checks the syntax of an agent key.
func ValidateAgentKey(key string) error {
	if !reAgentKey.MatchString(key) {
		return errors.New("invalid agent key format: must match ^[a-z][a-z0-9_]*$")
	}
	return nil
}

// ValidateChange validates a proposed ChangeRequest against domain rules and types.
func ValidateChange(req ChangeRequest) error {
	switch req.Operation {
	case "setting":
		return validateSettingChange(req)
	case "agent_config":
		return validateAgentConfigChange(req)
	case "agent_prompt":
		return validateAgentPromptChange(req)
	case "profile_activate":
		return validateProfileActivateChange(req)
	case "task_control":
		return validateTaskControlChange(req)
	default:
		return fmt.Errorf("unsupported operation: %s", req.Operation)
	}
}

func validateSettingChange(req ChangeRequest) error {
	if req.Target != "" {
		return errors.New("setting target must be empty")
	}
	if req.Key == "task_concurrency_enabled" || req.Key == "task_concurrency_limit" {
		return errors.New("concurrency settings are read-only in v1")
	}

	rawVal := strings.TrimSpace(string(req.Value))
	if rawVal == "" || rawVal == "null" {
		return errors.New("setting value cannot be empty or null")
	}

	switch req.Key {
	case "traffic_capture", "llm_record", "web_search_enabled", "llm_pool_enabled",
		"llm_pool_bind_fallback", "noa_compaction", "notify_enabled":
		if rawVal != "true" && rawVal != "false" {
			return errors.New("must be a boolean (true or false)")
		}
		return nil

	case "workers":
		var num json.Number
		dec := json.NewDecoder(strings.NewReader(rawVal))
		dec.UseNumber()
		if err := dec.Decode(&num); err != nil {
			return errors.New("workers must be an integer")
		}
		val, err := num.Int64()
		if err != nil || val <= 0 || val > math.MaxInt32 {
			return errors.New("workers must be a positive integer within 32-bit range")
		}
		return nil

	case "notify_digest_interval_min":
		var num json.Number
		dec := json.NewDecoder(strings.NewReader(rawVal))
		dec.UseNumber()
		if err := dec.Decode(&num); err != nil {
			return errors.New("notify_digest_interval_min must be an integer")
		}
		val, err := num.Int64()
		if err != nil || val < 1 || val > 1440 {
			return errors.New("notify_digest_interval_min must be between 1 and 1440")
		}
		return nil

	default:
		return fmt.Errorf("unsupported or read-only setting key: %s", req.Key)
	}
}

func validateAgentConfigChange(req ChangeRequest) error {
	if err := ValidateAgentKey(req.Target); err != nil {
		return err
	}

	rawVal := strings.TrimSpace(string(req.Value))
	if rawVal == "" {
		return errors.New("value cannot be empty")
	}

	switch req.Key {
	case "max_turns", "run_seconds":
		if rawVal == "null" {
			return fmt.Errorf("%s cannot be null", req.Key)
		}
		var num json.Number
		dec := json.NewDecoder(strings.NewReader(rawVal))
		dec.UseNumber()
		if err := dec.Decode(&num); err != nil {
			return fmt.Errorf("%s must be an integer", req.Key)
		}
		val, err := num.Int64()
		if err != nil {
			return fmt.Errorf("%s must be an integer", req.Key)
		}
		if val < 0 {
			return fmt.Errorf("%s must be non-negative", req.Key)
		}
		if val > math.MaxInt32 {
			return fmt.Errorf("%s exceeds 32-bit range", req.Key)
		}
		return nil

	case "web_search":
		if rawVal != "true" && rawVal != "false" {
			return errors.New("web_search must be a boolean (true or false)")
		}
		return nil

	case "llm_profile_id":
		if rawVal == "null" {
			return nil // null is explicit unbind
		}
		var str string
		if err := json.Unmarshal(req.Value, &str); err != nil {
			return errors.New("profile ID must be a positive decimal string or null")
		}
		if err := ValidateID(str); err != nil {
			if strings.Contains(err.Error(), "exceeds 64-bit range") {
				return errors.New("profile ID exceeds 64-bit range")
			}
			return fmt.Errorf("profile ID must be a positive decimal string: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported agent_config key: %s", req.Key)
	}
}

func validateAgentPromptChange(req ChangeRequest) error {
	if err := ValidateAgentKey(req.Target); err != nil {
		return err
	}
	if req.Key != "" {
		return errors.New("key must be empty for agent_prompt operation")
	}

	type promptPayload struct {
		Template string `json:"template"`
		Note     string `json:"note"`
	}

	var p promptPayload
	dec := json.NewDecoder(bytes.NewReader(req.Value))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return fmt.Errorf("unknown field in prompt payload: %w", err)
	}

	if p.Template == "" {
		return errors.New("template cannot be empty")
	}
	if len(p.Template) > 65536 {
		return errors.New("template exceeds 65536 bytes")
	}
	if len(p.Note) > 256 {
		return errors.New("note exceeds 256 bytes")
	}

	return nil
}

func validateProfileActivateChange(req ChangeRequest) error {
	if err := ValidateID(req.Target); err != nil {
		return fmt.Errorf("profile ID must be a positive decimal string: %w", err)
	}
	if req.Key != "" {
		return errors.New("key must be empty for profile_activate")
	}
	if len(req.Value) > 0 && string(req.Value) != "null" {
		return errors.New("value must be empty for profile_activate")
	}
	return nil
}

func validateTaskControlChange(req ChangeRequest) error {
	if err := ValidateID(req.Target); err != nil {
		return fmt.Errorf("task ID must be a positive decimal string: %w", err)
	}
	if req.Key != "pause" && req.Key != "resume" {
		return errors.New("task_control key must be pause or resume")
	}
	if len(req.Value) > 0 && string(req.Value) != "null" {
		return errors.New("value must be empty for task_control")
	}
	return nil
}
