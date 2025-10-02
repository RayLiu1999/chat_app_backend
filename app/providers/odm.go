package providers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// odm 錯誤定義
var (
	ErrInvalidModel     = errors.New("無效的模型結構")
	ErrDocumentNotFound = errors.New("文檔不存在")
	ErrInvalidID        = errors.New("無效的ID格式")
	ErrNoDocumentID     = errors.New("文檔沒有ID欄位")
)

// Model 介面定義所有模型必須實現的方法
type Model interface {
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
	GetCollectionName() string
}

// BaseModel 提供基本模型功能的實現
type BaseModel struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// GetID 獲取文檔ID
func (m *BaseModel) GetID() primitive.ObjectID {
	return m.ID
}

// SetID 設置文檔ID
func (m *BaseModel) SetID(id primitive.ObjectID) {
	m.ID = id
}

// GetBaseModel 返回 BaseModel 指標
func (m *BaseModel) GetBaseModel() *BaseModel {
	return m
}

// setTimestamps 設定模型的時間戳
func setTimestamps(model Model, isCreate bool) {
	// 嘗試通過 BaseModel 欄位設定時間
	if modelValue := reflect.ValueOf(model).Elem(); modelValue.IsValid() {
		if baseField := modelValue.FieldByName("BaseModel"); baseField.IsValid() && baseField.CanAddr() {
			now := time.Now()
			baseModel := baseField.Addr().Interface().(*BaseModel)
			if isCreate && baseModel.CreatedAt.IsZero() {
				baseModel.CreatedAt = now
			}
			baseModel.UpdatedAt = now
			return
		}
	}

	// 嘗試通過 GetBaseModel 方法設定時間
	if baseModel, ok := model.(interface{ GetBaseModel() *BaseModel }); ok {
		now := time.Now()
		base := baseModel.GetBaseModel()
		if isCreate && base.CreatedAt.IsZero() {
			base.CreatedAt = now
		}
		base.UpdatedAt = now
	}
}

// odm 提供對模型的資料庫操作
type odm struct {
	db *mongo.Database
}

// NewODM 創建新的ODM實例
func NewODM(db *mongo.Database) *odm {
	return &odm{
		db: db,
	}
}

// ===== 基礎工具方法 =====

// GetDatabase 返回數據庫連接
func (o *odm) GetDatabase() *mongo.Database {
	return o.db
}

// Collection 獲取模型對應的集合
func (o *odm) Collection(model Model) *mongo.Collection {
	return o.db.Collection(model.GetCollectionName())
}

// ===== 創建操作 =====

// Create 創建新文檔
func (o *odm) Create(ctx context.Context, model Model) error {
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		return ErrInvalidModel
	}

	// 設置創建和更新時間
	setTimestamps(model, true)

	// 如果ID為空，則生成新ID
	if model.GetID().IsZero() {
		model.SetID(primitive.NewObjectID())
	}

	_, err := o.Collection(model).InsertOne(ctx, model)
	return err
}

// InsertMany 插入多個文檔
func (o *odm) InsertMany(ctx context.Context, models []Model) error {
	if len(models) == 0 {
		return nil
	}

	// 將 []Model 轉換為 []any
	interfaces := make([]any, len(models))
	for i, model := range models {
		// 設定創建和更新時間
		setTimestamps(model, true)

		// 如果ID為空，則生成新ID
		if model.GetID().IsZero() {
			model.SetID(primitive.NewObjectID())
		}

		interfaces[i] = model
	}

	_, err := o.Collection(models[0]).InsertMany(ctx, interfaces)
	return err
}

// ===== 查詢操作 =====

// FindByID 通過ID查找文檔
func (o *odm) FindByID(ctx context.Context, ID string, model Model) error {
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		return ErrInvalidModel
	}

	objectID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return ErrInvalidID
	}

	filter := bson.M{"_id": objectID}
	err = o.Collection(model).FindOne(ctx, filter).Decode(model)
	fmt.Printf("Error: %v\n", err)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrDocumentNotFound
		}
		return err
	}

	return nil
}

// FindOne 查找單個文檔
func (o *odm) FindOne(ctx context.Context, filter bson.M, model Model) error {
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		return ErrInvalidModel
	}

	err := o.Collection(model).FindOne(ctx, filter).Decode(model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrDocumentNotFound
		}
		return err
	}

	return nil
}

// Find 查找多個文檔
func (o *odm) Find(ctx context.Context, filter bson.M, models any) error {
	// 確保models是指向切片的指針
	modelsValue := reflect.ValueOf(models)
	if modelsValue.Kind() != reflect.Ptr || modelsValue.Elem().Kind() != reflect.Slice {
		return ErrInvalidModel
	}

	// 獲取切片元素類型
	sliceValue := modelsValue.Elem()
	elemType := sliceValue.Type().Elem()

	// 創建一個新的實例來獲取集合名稱
	modelInstance := reflect.New(elemType).Interface()
	model, ok := modelInstance.(Model)
	if !ok {
		return ErrInvalidModel
	}

	cursor, err := o.Collection(model).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, models)
}

// ===== 更新操作 =====

// Update 更新文檔
func (o *odm) Update(ctx context.Context, model Model) error {
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		return ErrInvalidModel
	}

	id := model.GetID()
	if id.IsZero() {
		return ErrNoDocumentID
	}

	// 更新更新時間
	setTimestamps(model, false)

	filter := bson.M{"_id": id}
	_, err := o.Collection(model).ReplaceOne(ctx, filter, model)
	return err
}

// UpdateFields 更新文檔的特定欄位
func (o *odm) UpdateFields(ctx context.Context, model Model, fields bson.M) error {
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		return ErrInvalidModel
	}

	id := model.GetID()
	if id.IsZero() {
		return ErrNoDocumentID
	}

	// 添加更新時間
	fields["updated_at"] = time.Now()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": fields}
	_, err := o.Collection(model).UpdateOne(ctx, filter, update)
	return err
}

// UpdateMany 更新多個文檔
func (o *odm) UpdateMany(ctx context.Context, model Model, filter bson.M, update bson.M) error {
	// 添加更新時間
	if updateSet, ok := update["$set"]; ok {
		if updateSetMap, ok := updateSet.(bson.M); ok {
			updateSetMap["updated_at"] = time.Now()
		}
	} else {
		update["$set"] = bson.M{"updated_at": time.Now()}
	}

	_, err := o.Collection(model).UpdateMany(ctx, filter, update)
	return err
}

// ===== 刪除操作 =====

// Delete 刪除文檔
func (o *odm) Delete(ctx context.Context, model Model) error {
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		return ErrInvalidModel
	}

	id := model.GetID()
	if id.IsZero() {
		return ErrNoDocumentID
	}

	filter := bson.M{"_id": id}
	_, err := o.Collection(model).DeleteOne(ctx, filter)
	return err
}

// DeleteMany 刪除多個文檔
func (o *odm) DeleteMany(ctx context.Context, model Model, filter bson.M) error {
	_, err := o.Collection(model).DeleteMany(ctx, filter)
	return err
}

// DeleteByID 通過ID刪除文檔
func (o *odm) DeleteByID(ctx context.Context, ID string, model Model) error {
	objectID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return ErrInvalidID
	}
	filter := bson.M{"_id": objectID}
	_, err = o.Collection(model).DeleteOne(ctx, filter)
	return err
}

// ===== 統計和工具方法 =====

// Count 計算符合條件的文檔數量
func (o *odm) Count(ctx context.Context, filter bson.M, model Model) (int64, error) {
	return o.Collection(model).CountDocuments(ctx, filter)
}

// ===== 存在性檢查 =====

// Exists 檢查文檔是否存在
func (o *odm) Exists(ctx context.Context, filter bson.M, model Model) (bool, error) {
	count, err := o.Count(ctx, filter, model)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByID 通過ID檢查文檔是否存在
func (o *odm) ExistsByID(ctx context.Context, ID string, model Model) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return false, ErrInvalidID
	}
	filter := bson.M{"_id": objectID}
	return o.Exists(ctx, filter, model)
}

// ===== 高級查詢操作 =====

// FindWithOptions 使用自定義選項查找文檔
func (o *odm) FindWithOptions(ctx context.Context, filter bson.M, models any, options *QueryOptions) error {
	// 確保models是指向切片的指針
	modelsValue := reflect.ValueOf(models)
	if modelsValue.Kind() != reflect.Ptr || modelsValue.Elem().Kind() != reflect.Slice {
		return ErrInvalidModel
	}

	// 獲取切片元素類型
	sliceValue := modelsValue.Elem()
	elemType := sliceValue.Type().Elem()

	// 創建一個新的實例來獲取集合名稱
	modelInstance := reflect.New(elemType).Interface()
	model, ok := modelInstance.(Model)
	if !ok {
		return ErrInvalidModel
	}

	// 構建MongoDB選項
	findOptions := options.ToFindOptions()

	cursor, err := o.Collection(model).Find(ctx, filter, findOptions)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, models)
}

// Aggregate 執行聚合查詢
func (o *odm) Aggregate(ctx context.Context, pipeline any, models any, model Model) error {
	cursor, err := o.Collection(model).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, models)
}

// ===== 批量操作 =====

// BulkWrite 執行批量寫入操作
func (o *odm) BulkWrite(ctx context.Context, operations []mongo.WriteModel, model Model) (*mongo.BulkWriteResult, error) {
	return o.Collection(model).BulkWrite(ctx, operations)
}

// ===== 查詢選項工具 =====

// QueryOptions 定義查詢選項
type QueryOptions struct {
	Sort       bson.D
	Limit      *int64
	Skip       *int64
	Projection bson.M
}

// ToFindOptions 將QueryOptions轉換為MongoDB的FindOptions
func (qo *QueryOptions) ToFindOptions() *options.FindOptions {
	opts := options.Find()

	if len(qo.Sort) > 0 {
		opts.SetSort(qo.Sort)
	}

	if qo.Limit != nil {
		opts.SetLimit(*qo.Limit)
	}

	if qo.Skip != nil {
		opts.SetSkip(*qo.Skip)
	}

	if len(qo.Projection) > 0 {
		opts.SetProjection(qo.Projection)
	}

	return opts
}
