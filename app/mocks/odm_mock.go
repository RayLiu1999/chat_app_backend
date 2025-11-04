package mocks

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"context"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ODM 是 providers.ODM 介面的 mock 實作
type ODM struct {
	mock.Mock
}

func (m *ODM) GetDatabase() *mongo.Database {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*mongo.Database)
}

func (m *ODM) Collection(model providers.Model) *mongo.Collection {
	args := m.Called(model)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*mongo.Collection)
}

func (m *ODM) Create(ctx context.Context, model providers.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *ODM) InsertMany(ctx context.Context, models []providers.Model) error {
	args := m.Called(ctx, models)
	return args.Error(0)
}

func (m *ODM) FindByID(ctx context.Context, ID string, model providers.Model) error {
	args := m.Called(ctx, ID, model)
	return args.Error(0)
}

func (m *ODM) FindOne(ctx context.Context, filter bson.M, result providers.Model) error {
	args := m.Called(ctx, filter, result)

	// 自動填充結果
	if args.Get(0) != nil {
		switch r := result.(type) {
		case *models.DMRoom:
			*r = args.Get(0).(models.DMRoom)
		case *models.Channel:
			*r = args.Get(0).(models.Channel)
		case *models.Server:
			*r = args.Get(0).(models.Server)
		case *models.User:
			*r = args.Get(0).(models.User)
		case *models.Message:
			*r = args.Get(0).(models.Message)
		}
	}

	return args.Error(1)
}

func (m *ODM) Find(ctx context.Context, filter bson.M, models any) error {
	args := m.Called(ctx, filter, models)
	return args.Error(0)
}

func (m *ODM) Update(ctx context.Context, model providers.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *ODM) UpdateFields(ctx context.Context, model providers.Model, fields bson.M) error {
	args := m.Called(ctx, model, fields)
	return args.Error(0)
}

func (m *ODM) UpdateMany(ctx context.Context, model providers.Model, filter bson.M, update bson.M) error {
	args := m.Called(ctx, model, filter, update)
	return args.Error(0)
}

func (m *ODM) Delete(ctx context.Context, model providers.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *ODM) DeleteMany(ctx context.Context, model providers.Model, filter bson.M) error {
	args := m.Called(ctx, model, filter)
	return args.Error(0)
}

func (m *ODM) DeleteByID(ctx context.Context, ID string, model providers.Model) error {
	args := m.Called(ctx, ID, model)
	return args.Error(0)
}

func (m *ODM) Count(ctx context.Context, filter bson.M, model providers.Model) (int64, error) {
	args := m.Called(ctx, filter, model)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ODM) Exists(ctx context.Context, filter bson.M, model providers.Model) (bool, error) {
	args := m.Called(ctx, filter, model)
	return args.Bool(0), args.Error(1)
}

func (m *ODM) ExistsByID(ctx context.Context, ID string, model providers.Model) (bool, error) {
	args := m.Called(ctx, ID, model)
	return args.Bool(0), args.Error(1)
}

func (m *ODM) FindWithOptions(ctx context.Context, filter bson.M, models any, options *providers.QueryOptions) error {
	args := m.Called(ctx, filter, models, options)
	return args.Error(0)
}

func (m *ODM) Aggregate(ctx context.Context, pipeline any, models any, model providers.Model) error {
	args := m.Called(ctx, pipeline, models, model)
	return args.Error(0)
}

func (m *ODM) BulkWrite(ctx context.Context, operations []mongo.WriteModel, model providers.Model) (*mongo.BulkWriteResult, error) {
	args := m.Called(ctx, operations, model)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.BulkWriteResult), args.Error(1)
}
