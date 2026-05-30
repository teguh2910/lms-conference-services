package route

import (
	"database/sql"
	"log"

	"google.golang.org/grpc"

	conferencesDomain "lms-conference-service/internal/domain/conferences"
	"lms-conference-service/internal/pkg/db/redis"
	conferencesPb "lms-conference-service/pb/conferences"
)

// GrpcRoute func
func GrpcRoute(grpcServer *grpc.Server, db *sql.DB, log *log.Logger, cache *redis.Cache) {
	// Conference service
	conferenceServer := conferencesDomain.ConferenceServiceServer{Db: db, Cache: cache, Log: log}
	conferencesPb.RegisterConferenceServiceServer(grpcServer, &conferenceServer)

	// Conference participant service
	participantServer := conferencesDomain.ConferenceParticipantServiceServer{Db: db, Cache: cache, Log: log}
	conferencesPb.RegisterConferenceParticipantServiceServer(grpcServer, &participantServer)
}
