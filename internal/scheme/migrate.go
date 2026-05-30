package scheme

import (
	"database/sql"

	"github.com/GuiaBolso/darwin"
)

var migrations = []darwin.Migration{
	{
		Version:     1,
		Description: "Create uuid extension",
		Script:      `CREATE EXTENSION "uuid-ossp";`,
	},
	{
		Version:     2,
		Description: "Create conferences table",
		Script: `
			CREATE TABLE conferences (
				id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
				subject_class_id UUID NOT NULL,
				topic_subject_id UUID,
				name VARCHAR(100) NOT NULL,
				description TEXT,
				start_time TIMESTAMPTZ NOT NULL,
				end_time TIMESTAMPTZ NOT NULL,
				meeting_url VARCHAR(255),
				status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
				updated_by UUID,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', NOW()),
				created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', NOW()),
				deleted_at TIMESTAMPTZ
			);
		`,
	},
	{
		Version:     3,
		Description: "Create student_conferences table",
		Script: `
			CREATE TABLE student_conferences (
				id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
				conference_id UUID NOT NULL,
				student_id UUID NOT NULL,
				student_name VARCHAR(45) NOT NULL,
				is_attended BOOLEAN DEFAULT FALSE,
				joined_at TIMESTAMPTZ,
				left_at TIMESTAMPTZ,
				created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', NOW()),
				deleted_at TIMESTAMPTZ,
				FOREIGN KEY (conference_id) REFERENCES conferences(id) ON UPDATE CASCADE ON DELETE CASCADE
			);
		`,
	},
}

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sql.DB) error {
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	return d.Migrate()
}
