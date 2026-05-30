package conferences

import (
	"context"
	"database/sql"
	"log"
	"time"

	"lms-conference-service/internal/pkg/app"
	"lms-conference-service/internal/pkg/db/redis"
	conferencesPb "lms-conference-service/pb/conferences"
	genericPb "lms-conference-service/pb/generic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConferenceServiceServer struct
type ConferenceServiceServer struct {
	Db    *sql.DB
	Cache *redis.Cache
	Log   *log.Logger
	conferencesPb.UnimplementedConferenceServiceServer
}

// List conferences
func (s *ConferenceServiceServer) List(ctx context.Context, in *conferencesPb.ConferenceListInput) (*conferencesPb.ConferenceList, error) {
	repo := ConferenceRepository{Db: s.Db}

	pagination := in.GetPagination()
	limit := uint32(10)
	offset := uint32(0)
	keyword := ""
	orderBy := "created_at"
	sort := "DESC"

	if pagination != nil {
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
		offset = pagination.Offset
		if pagination.Keyword != "" {
			keyword = pagination.Keyword
		}
		if pagination.Order != "" {
			orderBy = pagination.Order
		}
		if pagination.Sort != "" {
			sort = pagination.Sort
		}
	}

	conferences, count, err := repo.List(ctx, in.GetSubjectClassId(), limit, offset, keyword, orderBy, sort)
	if err != nil {
		s.Log.Printf("error listing conferences: %v", err)
		return nil, status.Error(codes.Internal, "failed to list conferences")
	}

	return &conferencesPb.ConferenceList{
		Conferences: conferences,
		Count:       count,
	}, nil
}

// Get conference by ID (Issue #4)
func (s *ConferenceServiceServer) Get(ctx context.Context, in *genericPb.Id) (*conferencesPb.Conference, error) {
	repo := ConferenceRepository{Db: s.Db}

	conference, err := repo.Get(ctx, in.GetId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "conference not found")
		}
		s.Log.Printf("error getting conference: %v", err)
		return nil, status.Error(codes.Internal, "failed to get conference")
	}

	return conference, nil
}

// Create conference (Issue #1)
func (s *ConferenceServiceServer) Create(ctx context.Context, in *conferencesPb.ConferenceInput) (*conferencesPb.Conference, error) {
	repo := ConferenceRepository{Db: s.Db}

	userID := ctx.Value(app.Ctx("user_id")).(string)

	// Validation
	if in.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if in.GetSubjectClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "subject_class_id is required")
	}
	if in.GetStartTime() == "" {
		return nil, status.Error(codes.InvalidArgument, "start_time is required")
	}
	if in.GetEndTime() == "" {
		return nil, status.Error(codes.InvalidArgument, "end_time is required")
	}

	conference := &conferencesPb.Conference{
		SubjectClassId: in.GetSubjectClassId(),
		TopicSubjectId: in.GetTopicSubjectId(),
		Name:           in.GetName(),
		Description:    in.GetDescription(),
		StartTime:      in.GetStartTime(),
		EndTime:        in.GetEndTime(),
		MeetingUrl:     in.GetMeetingUrl(),
		UpdatedBy:      userID,
	}

	err := repo.Create(ctx, conference)
	if err != nil {
		s.Log.Printf("error creating conference: %v", err)
		return nil, status.Error(codes.Internal, "failed to create conference")
	}

	return conference, nil
}

// Update conference (Issue #2)
func (s *ConferenceServiceServer) Update(ctx context.Context, in *conferencesPb.Conference) (*conferencesPb.Conference, error) {
	repo := ConferenceRepository{Db: s.Db}

	userID := ctx.Value(app.Ctx("user_id")).(string)

	// Validation
	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if in.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	in.UpdatedBy = userID
	in.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	err := repo.Update(ctx, in)
	if err != nil {
		s.Log.Printf("error updating conference: %v", err)
		return nil, status.Error(codes.Internal, "failed to update conference")
	}

	return in, nil
}

// Delete conference (Issue #3) - soft delete
func (s *ConferenceServiceServer) Delete(ctx context.Context, in *genericPb.Id) (*genericPb.BoolMessage, error) {
	repo := ConferenceRepository{Db: s.Db}

	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := repo.Delete(ctx, in.GetId())
	if err != nil {
		s.Log.Printf("error deleting conference: %v", err)
		return &genericPb.BoolMessage{IsTrue: false}, status.Error(codes.Internal, "failed to delete conference")
	}

	return &genericPb.BoolMessage{IsTrue: true}, nil
}

// ConferenceParticipantServiceServer struct (handles Join Conference - Issue #5)
type ConferenceParticipantServiceServer struct {
	Db    *sql.DB
	Cache *redis.Cache
	Log   *log.Logger
	conferencesPb.UnimplementedConferenceParticipantServiceServer
}

// List participants
func (s *ConferenceParticipantServiceServer) List(ctx context.Context, in *conferencesPb.ConferenceParticipantListInput) (*conferencesPb.ConferenceParticipantList, error) {
	repo := StudentConferenceRepository{Db: s.Db}

	pagination := in.GetPagination()
	limit := uint32(10)
	offset := uint32(0)
	keyword := ""
	orderBy := "created_at"
	sort := "DESC"

	if pagination != nil {
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
		offset = pagination.Offset
		if pagination.Keyword != "" {
			keyword = pagination.Keyword
		}
		if pagination.Order != "" {
			orderBy = pagination.Order
		}
		if pagination.Sort != "" {
			sort = pagination.Sort
		}
	}

	participants, count, err := repo.List(ctx, in.GetConferenceId(), limit, offset, keyword, orderBy, sort)
	if err != nil {
		s.Log.Printf("error listing participants: %v", err)
		return nil, status.Error(codes.Internal, "failed to list participants")
	}

	return &conferencesPb.ConferenceParticipantList{
		Participants: participants,
		Count:        count,
	}, nil
}

// Create - Join Conference (Issue #5)
// When a student joins a conference, insert into student_conferences table
func (s *ConferenceParticipantServiceServer) Create(ctx context.Context, in *conferencesPb.ConferenceParticipantInput) (*conferencesPb.ConferenceParticipant, error) {
	repo := StudentConferenceRepository{Db: s.Db}

	// Validation
	if in.GetConferenceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "conference_id is required")
	}
	if in.GetStudentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "student_id is required")
	}
	if in.GetStudentName() == "" {
		return nil, status.Error(codes.InvalidArgument, "student_name is required")
	}

	participant := &conferencesPb.ConferenceParticipant{
		ConferenceId: in.GetConferenceId(),
		StudentId:    in.GetStudentId(),
		StudentName:  in.GetStudentName(),
	}

	err := repo.Create(ctx, participant)
	if err != nil {
		s.Log.Printf("error joining conference: %v", err)
		return nil, status.Error(codes.Internal, "failed to join conference")
	}

	return participant, nil
}

// UpdateAttendance updates participant attendance
func (s *ConferenceParticipantServiceServer) UpdateAttendance(ctx context.Context, in *conferencesPb.UpdateAttendanceInput) (*conferencesPb.ConferenceParticipant, error) {
	repo := StudentConferenceRepository{Db: s.Db}

	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	participant := &conferencesPb.ConferenceParticipant{
		Id:         in.GetId(),
		IsAttended: in.GetIsAttended(),
		JoinedAt:   in.GetJoinedAt(),
		LeftAt:     in.GetLeftAt(),
	}

	err := repo.UpdateAttendance(ctx, participant)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		s.Log.Printf("error updating attendance: %v", err)
		return nil, status.Error(codes.Internal, "failed to update attendance")
	}

	return participant, nil
}

// Delete participant
func (s *ConferenceParticipantServiceServer) Delete(ctx context.Context, in *genericPb.Id) (*genericPb.BoolMessage, error) {
	repo := StudentConferenceRepository{Db: s.Db}

	err := repo.Delete(ctx, in.GetId())
	if err != nil {
		s.Log.Printf("error deleting participant: %v", err)
		return &genericPb.BoolMessage{IsTrue: false}, status.Error(codes.Internal, "failed to delete participant")
	}

	return &genericPb.BoolMessage{IsTrue: true}, nil
}
