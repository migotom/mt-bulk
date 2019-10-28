package vulnerabilities

import (
	"go.uber.org/zap"
)

type dbLog struct {
	Sugar *zap.SugaredLogger
}

func (l *dbLog) Errorf(f string, v ...interface{}) {
	l.Sugar.Errorf(f, v...)
}

func (l *dbLog) Warningf(f string, v ...interface{}) {
	l.Sugar.Warnf(f, v...)
}

func (l *dbLog) Infof(f string, v ...interface{}) {
}

func (l *dbLog) Debugf(f string, v ...interface{}) {
}
