package mongo

import (
	"context"

	"github.com/sirupsen/logrus"
	logutil "gitscm.cisco.com/mcmp/utils/log"
)

func log(ctx context.Context) logrus.FieldLogger {
	return logutil.Logger(ctx).WithField("pkg", "mongo")
}
