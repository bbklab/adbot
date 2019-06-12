package scheduler

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/bbklab/adbot/types"
)

var (
	formatPlain = "plain"
	formatJSON  = "json"
)

type auditLogger struct {
	w      io.Writer
	format string
}

func newRollingAuditLogger() *auditLogger {
	writer := &lumberjack.Logger{
		Filename:   "/var/log/adbot-audit/adbot-audit",
		MaxSize:    100,
		MaxBackups: 0,   // retain all, ignore number of files
		MaxAge:     365, // retain log files in one year
	}

	return newAuditLogger(writer, formatJSON)
}

func newAuditLogger(w io.Writer, format string) *auditLogger {
	if format == "" {
		format = formatPlain
	}
	return &auditLogger{
		w:      w,
		format: format,
	}
}

func (a *auditLogger) logEntry(es ...*types.AuditEntry) {
	for _, e := range es {
		var line string

		switch a.format {
		case formatPlain:
			line = e.FormatString() + "\n"
		case formatJSON:
			line = e.FormatJSON() + "\n"
		}

		if _, err := fmt.Fprint(a.w, line); err != nil {
			log.Errorln("audit.logEntry() error:", err)
		}
	}
}

// LogAuditEntry is exported to log audit entries
func LogAuditEntry(es ...*types.AuditEntry) {
	sched.auditLogger.logEntry(es...)
}
