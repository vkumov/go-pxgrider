package connection

import (
	"context"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vkumov/go-pxgrider/server/internal/db/models"
)

type LogsSlice models.LogSlice

func (c *Connection) GetLogs(ctx context.Context, limit, offset int64) (LogsSlice, error) {
	q := []qm.QueryMod{
		models.LogWhere.Client.EQ(c.id),
		qm.OrderBy(models.LogColumns.ID + " DESC"),
	}

	if limit > 0 {
		q = append(q, qm.Limit(int(limit)))
		q = append(q, qm.Offset(int(offset)))
	}

	raw, err := models.Logs(q...).All(ctx, c.db.Load())
	if err != nil {
		return nil, err
	}

	return LogsSlice(raw), nil
}

func (c *Connection) GetLogsCount(ctx context.Context) (int64, error) {
	return models.Logs(models.LogWhere.Client.EQ(c.id)).Count(ctx, c.db.Load())
}

func (m LogsSlice) ToProto() []*pb.ConnectionLog {
	var res []*pb.ConnectionLog

	for _, v := range m {
		var ts *timestamppb.Timestamp
		if v.Timestamp.Valid {
			ts = timestamppb.New(v.Timestamp.Time)
		}

		res = append(res, &pb.ConnectionLog{
			Id:        v.ID,
			Client:    v.Client,
			Level:     v.Level,
			Timestamp: ts,
			Message:   v.Message.String,
			Label:     v.Label.String,
		})
	}

	return res
}

func (c *Connection) DeleteAllLogs(ctx context.Context) (int64, error) {
	return models.Logs(models.LogWhere.Client.EQ(c.id)).DeleteAll(ctx, c.db.Load())
}

func (c *Connection) DeleteLogs(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	return models.Logs(
		models.LogWhere.ID.IN(ids),
		models.LogWhere.Client.EQ(c.id),
	).DeleteAll(ctx, c.db.Load())
}
