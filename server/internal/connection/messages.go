package connection

import (
	"context"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vkumov/go-pxgrider/server/internal/db/models"
)

type MessageSlice models.MessageSlice

func (c *Connection) GetMessages(ctx context.Context, limit, offset int64) (MessageSlice, error) {
	q := []qm.QueryMod{
		models.MessageWhere.Client.EQ(c.id),
		qm.OrderBy(models.MessageColumns.ID + " DESC"),
	}

	if limit > 0 {
		q = append(q, qm.Limit(int(limit)))
		q = append(q, qm.Offset(int(offset)))
	}

	raw, err := models.Messages(q...).All(ctx, c.db.Load())
	if err != nil {
		return nil, err
	}

	return MessageSlice(raw), nil
}

func (c *Connection) GetMessagesCount(ctx context.Context) (int64, error) {
	return models.Messages(models.MessageWhere.Client.EQ(c.id)).Count(ctx, c.db.Load())
}

func (m MessageSlice) ToProto() []*pb.ConnectionMessage {
	var res []*pb.ConnectionMessage

	for _, v := range m {
		var ts *timestamppb.Timestamp
		if v.Timestamp.Valid {
			ts = timestamppb.New(v.Timestamp.Time)
		}

		res = append(res, &pb.ConnectionMessage{
			Id:        v.ID,
			Client:    v.Client,
			Topic:     v.Topic,
			Message:   string(v.Message.JSON),
			Timestamp: ts,
			Viewed:    v.Viewed.Bool,
		})
	}

	return res
}

func (c *Connection) MarkMessages(ctx context.Context, ids []int64, viewed bool) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := models.Messages(
		models.MessageWhere.ID.IN(ids),
		models.MessageWhere.Client.EQ(c.id),
	).UpdateAll(ctx, c.db.Load(), models.M{models.MessageColumns.Viewed: viewed})
	return err
}

func (c *Connection) DeleteAllMessages(ctx context.Context) (int64, error) {
	return models.Messages(models.MessageWhere.Client.EQ(c.id)).DeleteAll(ctx, c.db.Load())
}

func (c *Connection) DeleteMessages(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	return models.Messages(
		models.MessageWhere.ID.IN(ids),
		models.MessageWhere.Client.EQ(c.id),
	).DeleteAll(ctx, c.db.Load())
}
