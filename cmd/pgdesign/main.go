package main

import (
	"fmt"

	"github.com/smm-h/strictcli/go/strictcli"
)

func main() {
	app := strictcli.NewApp("pgdesign", Version, "PostgreSQL schema compiler")

	app.GlobalFlag(strictcli.BoolFlag("quiet", "Suppress non-error output"))
	app.GlobalFlag(strictcli.StringFlag("db", "PostgreSQL connection URL", strictcli.Default(nil)))
	app.GlobalFlag(strictcli.BoolFlag("strict-nf", "Enable strict normal form checking"))

	app.Command("generate", "Generate SQL from a pgdesign schema file", notImplemented,
		strictcli.WithArgs(strictcli.NewArg("file", "Path to schema file")),
	)

	app.Command("validate", "Validate a pgdesign schema file", notImplemented,
		strictcli.WithArgs(strictcli.NewArg("file", "Path to schema file")),
	)

	app.Command("audit", "Audit a pgdesign schema file for issues", notImplemented,
		strictcli.WithArgs(strictcli.NewArg("file", "Path to schema file")),
	)

	app.Command("fmt", "Format a pgdesign schema file or directory", notImplemented,
		strictcli.WithArgs(strictcli.NewArg("path", "Path to file or directory")),
	)

	app.Command("introspect", "Introspect a live PostgreSQL database", notImplemented)

	app.Command("diff", "Diff a schema file against a live database", notImplemented,
		strictcli.WithArgs(strictcli.NewArg("file", "Path to schema file")),
	)

	migrate := app.Group("migrate", "Database migration commands")
	migrate.Command("plan", "Plan migrations from schema changes", notImplemented)
	migrate.Command("generate", "Generate migration files", notImplemented)
	migrate.Command("apply", "Apply pending migrations", notImplemented)
	migrate.Command("rollback", "Rollback the last migration", notImplemented)
	migrate.Command("status", "Show migration status", notImplemented)

	app.Command("serve", "Start the pgdesign language server", notImplemented)

	app.Run()
}

func notImplemented(_ map[string]interface{}) int {
	fmt.Println("not implemented")
	return 0
}
