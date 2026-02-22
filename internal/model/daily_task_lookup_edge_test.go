package model

import "testing"

func TestGetTaskConfig_UnknownTypeReturnsNil(t *testing.T) {
	unknown := DailyTaskType("not_defined")
	if cfg := GetTaskConfig(unknown); cfg != nil {
		t.Fatalf("expected nil config for unknown task type, got %+v", cfg)
	}
}

func TestPopulateTaskInfo_UnknownTypeKeepsVirtualFieldsEmpty(t *testing.T) {
	task := &DailyTask{TaskType: DailyTaskType("unknown")}
	task.PopulateTaskInfo()

	if task.TaskName != "" || task.TaskDescription != "" || task.TaskIcon != "" {
		t.Fatalf("expected virtual fields to remain empty for unknown task type, got %+v", task)
	}
}
