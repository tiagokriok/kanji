package repositories

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/tiagokriok/lazytask/internal/domain"
	"github.com/tiagokriok/lazytask/internal/infrastructure/db/sqlc"
)

func nullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func nullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullableTimeToString(v *time.Time) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: v.UTC().Format(time.RFC3339), Valid: true}
}

func parseRFC3339OrZero(in string) time.Time {
	if in == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, in)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func parseOptionalTime(in sql.NullString) *time.Time {
	if !in.Valid || in.String == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, in.String)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseLabels(raw string) []string {
	if raw == "" {
		return []string{}
	}
	labels := make([]string, 0)
	if err := json.Unmarshal([]byte(raw), &labels); err != nil {
		return []string{}
	}
	return labels
}

func marshalLabels(labels []string) string {
	if labels == nil {
		labels = []string{}
	}
	encoded, err := json.Marshal(labels)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func fromSQLTask(t sqlc.Task) domain.Task {
	var boardID *string
	if t.BoardID.Valid {
		boardID = &t.BoardID.String
	}
	var columnID *string
	if t.ColumnID.Valid {
		columnID = &t.ColumnID.String
	}
	var remoteID *string
	if t.RemoteID.Valid {
		remoteID = &t.RemoteID.String
	}
	var status *string
	if t.Status.Valid {
		status = &t.Status.String
	}
	var estimateMinutes *int
	if t.EstimateMinutes.Valid {
		value := int(t.EstimateMinutes.Int64)
		estimateMinutes = &value
	}
	var assignee *string
	if t.Assignee.Valid {
		assignee = &t.Assignee.String
	}

	return domain.Task{
		ID:              t.ID,
		ProviderID:      t.ProviderID,
		WorkspaceID:     t.WorkspaceID,
		BoardID:         boardID,
		ColumnID:        columnID,
		RemoteID:        remoteID,
		Title:           t.Title,
		DescriptionMD:   t.DescriptionMd,
		Status:          status,
		Priority:        int(t.Priority),
		DueAt:           parseOptionalTime(t.DueAt),
		EstimateMinutes: estimateMinutes,
		Assignee:        assignee,
		Labels:          parseLabels(t.LabelsJSON),
		Position:        t.Position,
		CreatedAt:       parseRFC3339OrZero(t.CreatedAt),
		UpdatedAt:       parseRFC3339OrZero(t.UpdatedAt),
	}
}

func fromSQLComment(c sqlc.Comment) domain.Comment {
	var remoteID *string
	if c.RemoteID.Valid {
		remoteID = &c.RemoteID.String
	}
	var author *string
	if c.Author.Valid {
		author = &c.Author.String
	}
	return domain.Comment{
		ID:         c.ID,
		TaskID:     c.TaskID,
		ProviderID: c.ProviderID,
		RemoteID:   remoteID,
		BodyMD:     c.BodyMd,
		Author:     author,
		CreatedAt:  parseRFC3339OrZero(c.CreatedAt),
	}
}

func fromSQLColumn(c sqlc.Column) domain.Column {
	var remoteID *string
	if c.RemoteID.Valid {
		remoteID = &c.RemoteID.String
	}
	var wipLimit *int
	if c.WipLimit.Valid {
		w := int(c.WipLimit.Int64)
		wipLimit = &w
	}
	return domain.Column{
		ID:       c.ID,
		BoardID:  c.BoardID,
		RemoteID: remoteID,
		Name:     c.Name,
		Position: int(c.Position),
		WIPLimit: wipLimit,
	}
}

func fromSQLBoard(b sqlc.Board) domain.Board {
	var remoteID *string
	if b.RemoteID.Valid {
		remoteID = &b.RemoteID.String
	}
	return domain.Board{
		ID:          b.ID,
		WorkspaceID: b.WorkspaceID,
		RemoteID:    remoteID,
		Name:        b.Name,
		ViewDefault: b.ViewDefault,
	}
}

func fromSQLWorkspace(w sqlc.Workspace) domain.Workspace {
	var remoteID *string
	if w.RemoteID.Valid {
		remoteID = &w.RemoteID.String
	}
	return domain.Workspace{
		ID:         w.ID,
		ProviderID: w.ProviderID,
		RemoteID:   remoteID,
		Name:       w.Name,
	}
}

func fromSQLProvider(p sqlc.Provider) domain.Provider {
	var auth *string
	if p.AuthJSON.Valid {
		auth = &p.AuthJSON.String
	}
	return domain.Provider{
		ID:        p.ID,
		Type:      p.Type,
		Name:      p.Name,
		AuthJSON:  auth,
		CreatedAt: parseRFC3339OrZero(p.CreatedAt),
	}
}
