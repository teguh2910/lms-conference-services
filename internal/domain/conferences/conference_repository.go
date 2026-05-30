package conferences

import (
	"context"
	"database/sql"
	"fmt"

	conferencesPb "lms-conference-service/pb/conferences"
)

// ConferenceRepository struct
type ConferenceRepository struct {
	Db *sql.DB
}

// List conferences with pagination
func (r *ConferenceRepository) List(ctx context.Context, subjectClassID string, limit, offset uint32, keyword, orderBy, sort string) ([]*conferencesPb.Conference, uint32, error) {
	query := `
		SELECT id, subject_class_id, COALESCE(topic_subject_id::text, ''), name, 
			COALESCE(description, ''), start_time, end_time, COALESCE(meeting_url, ''),
			status, COALESCE(updated_by::text, ''), updated_at, created_at
		FROM conferences
		WHERE subject_class_id = $1 AND name ILIKE $2 AND deleted_at IS NULL
		ORDER BY ` + orderBy + ` ` + sort + `
		LIMIT $3 OFFSET $4
	`

	searchKeyword := "%" + keyword + "%"
	rows, err := r.Db.QueryContext(ctx, query, subjectClassID, searchKeyword, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var conferences []*conferencesPb.Conference
	for rows.Next() {
		var c conferencesPb.Conference
		err := rows.Scan(
			&c.Id, &c.SubjectClassId, &c.TopicSubjectId, &c.Name,
			&c.Description, &c.StartTime, &c.EndTime, &c.MeetingUrl,
			&c.Status, &c.UpdatedBy, &c.UpdatedAt, &c.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		conferences = append(conferences, &c)
	}

	countQuery := `SELECT COUNT(*) FROM conferences WHERE subject_class_id = $1 AND name ILIKE $2 AND deleted_at IS NULL`
	var count uint32
	err = r.Db.QueryRowContext(ctx, countQuery, subjectClassID, searchKeyword).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return conferences, count, nil
}

// Get conference by ID
func (r *ConferenceRepository) Get(ctx context.Context, id string) (*conferencesPb.Conference, error) {
	query := `
		SELECT id, subject_class_id, COALESCE(topic_subject_id::text, ''), name,
			COALESCE(description, ''), start_time, end_time, COALESCE(meeting_url, ''),
			status, COALESCE(updated_by::text, ''), updated_at, created_at
		FROM conferences WHERE id = $1 AND deleted_at IS NULL
	`

	var c conferencesPb.Conference
	err := r.Db.QueryRowContext(ctx, query, id).Scan(
		&c.Id, &c.SubjectClassId, &c.TopicSubjectId, &c.Name,
		&c.Description, &c.StartTime, &c.EndTime, &c.MeetingUrl,
		&c.Status, &c.UpdatedBy, &c.UpdatedAt, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Create conference
func (r *ConferenceRepository) Create(ctx context.Context, c *conferencesPb.Conference) error {
	query := `
		INSERT INTO conferences (subject_class_id, topic_subject_id, name, description, start_time, end_time, meeting_url, updated_by)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7, $8)
		RETURNING id, status, updated_at, created_at
	`

	return r.Db.QueryRowContext(ctx, query,
		c.SubjectClassId, c.TopicSubjectId, c.Name, c.Description,
		c.StartTime, c.EndTime, c.MeetingUrl, c.UpdatedBy,
	).Scan(&c.Id, &c.Status, &c.UpdatedAt, &c.CreatedAt)
}

// Update conference
func (r *ConferenceRepository) Update(ctx context.Context, c *conferencesPb.Conference) error {
	query := `
		UPDATE conferences SET
			subject_class_id = $1, topic_subject_id = NULLIF($2, '')::uuid, name = $3,
			description = $4, start_time = $5, end_time = $6, meeting_url = $7,
			status = $8, updated_by = $9, updated_at = NOW()
		WHERE id = $10 AND deleted_at IS NULL
	`

	result, err := r.Db.ExecContext(ctx, query,
		c.SubjectClassId, c.TopicSubjectId, c.Name, c.Description,
		c.StartTime, c.EndTime, c.MeetingUrl, c.Status,
		c.UpdatedBy, c.Id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("conference not found")
	}

	return nil
}

// Delete conference (soft delete)
func (r *ConferenceRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE conferences SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("conference not found")
	}

	return nil
}

// StudentConferenceRepository struct
type StudentConferenceRepository struct {
	Db *sql.DB
}

// List student conferences (participants)
func (r *StudentConferenceRepository) List(ctx context.Context, conferenceID string, limit, offset uint32, keyword, orderBy, sort string) ([]*conferencesPb.ConferenceParticipant, uint32, error) {
	query := `
		SELECT id, conference_id, student_id, student_name, is_attended,
			COALESCE(joined_at::text, ''), COALESCE(left_at::text, ''), created_at
		FROM student_conferences
		WHERE conference_id = $1 AND student_name ILIKE $2 AND deleted_at IS NULL
		ORDER BY ` + orderBy + ` ` + sort + `
		LIMIT $3 OFFSET $4
	`

	searchKeyword := "%" + keyword + "%"
	rows, err := r.Db.QueryContext(ctx, query, conferenceID, searchKeyword, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var participants []*conferencesPb.ConferenceParticipant
	for rows.Next() {
		var p conferencesPb.ConferenceParticipant
		err := rows.Scan(
			&p.Id, &p.ConferenceId, &p.StudentId, &p.StudentName,
			&p.IsAttended, &p.JoinedAt, &p.LeftAt, &p.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		participants = append(participants, &p)
	}

	countQuery := `SELECT COUNT(*) FROM student_conferences WHERE conference_id = $1 AND student_name ILIKE $2 AND deleted_at IS NULL`
	var count uint32
	err = r.Db.QueryRowContext(ctx, countQuery, conferenceID, searchKeyword).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return participants, count, nil
}

// Create (Join conference) - insert student into student_conferences
func (r *StudentConferenceRepository) Create(ctx context.Context, p *conferencesPb.ConferenceParticipant) error {
	query := `
		INSERT INTO student_conferences (conference_id, student_id, student_name)
		VALUES ($1, $2, $3)
		RETURNING id, is_attended, created_at
	`

	return r.Db.QueryRowContext(ctx, query,
		p.ConferenceId, p.StudentId, p.StudentName,
	).Scan(&p.Id, &p.IsAttended, &p.CreatedAt)
}

// UpdateAttendance updates student attendance
func (r *StudentConferenceRepository) UpdateAttendance(ctx context.Context, p *conferencesPb.ConferenceParticipant) error {
	query := `
		UPDATE student_conferences SET
			is_attended = $1, joined_at = NULLIF($2, '')::timestamptz, left_at = NULLIF($3, '')::timestamptz
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING conference_id, student_id, student_name, created_at
	`

	return r.Db.QueryRowContext(ctx, query,
		p.IsAttended, p.JoinedAt, p.LeftAt, p.Id,
	).Scan(&p.ConferenceId, &p.StudentId, &p.StudentName, &p.CreatedAt)
}

// Delete student conference (soft delete)
func (r *StudentConferenceRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE student_conferences SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("student conference not found")
	}

	return nil
}
