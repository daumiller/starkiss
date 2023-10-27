package database

import (
  "database/sql"
)

type TranscodingTaskStatus string
const (
  TranscodingTaskTodo    TranscodingTaskStatus = "todo"
  TranscodingTaskRunning TranscodingTaskStatus = "running"
  TranscodingTaskSuccess TranscodingTaskStatus = "success"
  TranscodingTaskFailure TranscodingTaskStatus = "failure"
)
type TranscodingTask struct {
  Id            string                `json:"id"`
  UnprocessedId string                `json:"unprocessed_id"`
  TimeStarted   int64                 `json:"time_started"`
  TimeElapsed   int64                 `json:"time_elapsed"`
  Status        TranscodingTaskStatus `json:"status"`
  ErrorMessage  string                `json:"error_message"`
  CommandLine   string                `json:"command_line"`
}

// ============================================================================

func (tct *TranscodingTask) Create(db *sql.DB) (err error) {
  return TableCreate(db, tct)
}

func (tct *TranscodingTask) Update(db *sql.DB, new_values *TranscodingTask) (err error) {
  return TableUpdate(db, tct, new_values)
}

func (tct *TranscodingTask) Delete(db *sql.DB) (err error) {
  return TableDelete(db, tct)
}

// ============================================================================
// Table interface

func (tct *TranscodingTask) TableName() string { return "transcoding_tasks" }
func (tct *TranscodingTask) GetId() string { return tct.Id }
func (tct *TranscodingTask) SetId(id string) { tct.Id = id }

func (tct *TranscodingTask) CreateFrom(fields map[string]any) (instance Table, err error) {
  new_instance := TranscodingTask {}
  err = new_instance.FieldsWrite(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (tct *TranscodingTask) FieldsRead() (fields map[string]any, err error) {
  fields = make(map[string]any)
  fields["id"]                  = tct.Id
  fields["unprocessed_id"]      = tct.UnprocessedId
  fields["time_started"]        = tct.TimeStarted
  fields["time_elapsed"]        = tct.TimeElapsed
  fields["status"]              = tct.Status
  fields["error_message"]       = tct.ErrorMessage
  fields["command_line"]        = tct.CommandLine
  return fields, nil
}

func (tct *TranscodingTask) FieldsWrite(fields map[string]any) (err error) {
  tct.Id            = fields["id"].(string)
  tct.UnprocessedId = fields["unprocessed_id"].(string)
  tct.TimeStarted   = fields["time_started"].(int64)
  tct.TimeElapsed   = fields["time_elapsed"].(int64)
  tct.Status        = TranscodingTaskStatus(fields["status"].(string))
  tct.ErrorMessage  = fields["error_message"].(string)
  tct.CommandLine   = fields["command_line"].(string)
  return nil
}

func (tct_a *TranscodingTask) FieldsDifference(other Table) (diff map[string]any, err error) {
  diff = make(map[string]any)
  tct_b, b_is_tct := other.(*TranscodingTask)
  if b_is_tct == false { return diff, ErrInvalidType }

  if tct_b.Id            != tct_a.Id            { diff["id"]              = tct_b.Id            }
  if tct_b.UnprocessedId != tct_a.UnprocessedId { diff["unprocessed_id"]  = tct_b.UnprocessedId }
  if tct_b.TimeStarted   != tct_a.TimeStarted   { diff["time_started"]    = tct_b.TimeStarted   }
  if tct_b.TimeElapsed   != tct_a.TimeElapsed   { diff["time_elapsed"]    = tct_b.TimeElapsed   }
  if tct_b.Status        != tct_a.Status        { diff["status"]          = tct_b.Status        }
  if tct_b.ErrorMessage  != tct_a.ErrorMessage  { diff["error_message"]   = tct_b.ErrorMessage  }
  if tct_b.CommandLine   != tct_a.CommandLine   { diff["command_line"]    = tct_b.CommandLine   }

  return diff, nil
}

func (tct *TranscodingTask) ValidCreate(db *sql.DB) (valid bool, err error) {
  return true, nil
}

func (tct *TranscodingTask) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  return true, nil
}

func (tct *TranscodingTask) ValidDelete(db *sql.DB) (valid bool, err error) {
  return true, nil
}
