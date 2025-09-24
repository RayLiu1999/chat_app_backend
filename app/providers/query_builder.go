package providers

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QueryBuilder 提供流暢的 API 來構建 MongoDB 查詢
type QueryBuilder struct {
	filter  bson.M
	sort    bson.D
	limit   int64
	skip    int64
	project bson.M
}

// NewQueryBuilder 創建一個新的查詢構建器
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		filter:  bson.M{},
		sort:    bson.D{},
		limit:   0,
		skip:    0,
		project: bson.M{},
	}
}

// Where 添加一個等於條件
func (q *QueryBuilder) Where(field string, value any) *QueryBuilder {
	q.filter[field] = value
	return q
}

// WhereID 添加一個 ID 等於條件，當轉換失敗時返回錯誤
func (q *QueryBuilder) WhereID(id string) (*QueryBuilder, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return q, err
	}
	q.filter["_id"] = objectID
	return q, nil
}

// WhereIn 添加一個 $in 條件
func (q *QueryBuilder) WhereIn(field string, values any) *QueryBuilder {
	q.filter[field] = bson.M{"$in": values}
	return q
}

// WhereNotIn 添加一個 $nin 條件
func (q *QueryBuilder) WhereNotIn(field string, values any) *QueryBuilder {
	q.filter[field] = bson.M{"$nin": values}
	return q
}

// WhereGt 添加一個大於條件
func (q *QueryBuilder) WhereGt(field string, value any) *QueryBuilder {
	q.filter[field] = bson.M{"$gt": value}
	return q
}

// WhereGte 添加一個大於等於條件
func (q *QueryBuilder) WhereGte(field string, value any) *QueryBuilder {
	q.filter[field] = bson.M{"$gte": value}
	return q
}

// WhereLt 添加一個小於條件
func (q *QueryBuilder) WhereLt(field string, value any) *QueryBuilder {
	q.filter[field] = bson.M{"$lt": value}
	return q
}

// WhereLte 添加一個小於等於條件
func (q *QueryBuilder) WhereLte(field string, value any) *QueryBuilder {
	q.filter[field] = bson.M{"$lte": value}
	return q
}

// WhereNe 添加一個不等於條件
func (q *QueryBuilder) WhereNe(field string, value any) *QueryBuilder {
	q.filter[field] = bson.M{"$ne": value}
	return q
}

// WhereExists 添加一個存在條件
func (q *QueryBuilder) WhereExists(field string, exists bool) *QueryBuilder {
	q.filter[field] = bson.M{"$exists": exists}
	return q
}

// WhereRegex 添加一個正則表達式條件
func (q *QueryBuilder) WhereRegex(field string, pattern string, options string) *QueryBuilder {
	q.filter[field] = bson.M{"$regex": pattern, "$options": options}
	return q
}

// OrWhere 添加一個 $or 條件
func (q *QueryBuilder) OrWhere(conditions []bson.M) *QueryBuilder {
	q.filter["$or"] = conditions
	return q
}

// AndWhere 添加一個 $and 條件
func (q *QueryBuilder) AndWhere(conditions []bson.M) *QueryBuilder {
	q.filter["$and"] = conditions
	return q
}

// SortAsc 按欄位升序排序
func (q *QueryBuilder) SortAsc(field string) *QueryBuilder {
	q.sort = append(q.sort, bson.E{Key: field, Value: 1})
	return q
}

// SortDesc 按欄位降序排序
func (q *QueryBuilder) SortDesc(field string) *QueryBuilder {
	q.sort = append(q.sort, bson.E{Key: field, Value: -1})
	return q
}

// Limit 設置結果數量限制
func (q *QueryBuilder) Limit(limit int64) *QueryBuilder {
	q.limit = limit
	return q
}

// Skip 設置跳過的結果數量
func (q *QueryBuilder) Skip(skip int64) *QueryBuilder {
	q.skip = skip
	return q
}

// Select 選擇要包含的欄位
func (q *QueryBuilder) Select(fields ...string) *QueryBuilder {
	for _, field := range fields {
		q.project[field] = 1
	}
	return q
}

// Exclude 排除指定的欄位
func (q *QueryBuilder) Exclude(fields ...string) *QueryBuilder {
	for _, field := range fields {
		q.project[field] = 0
	}
	return q
}

// GetFilter 獲取構建的過濾條件
func (q *QueryBuilder) GetFilter() bson.M {
	return q.filter
}

// GetOptions 獲取構建的查詢選項
func (q *QueryBuilder) GetOptions() *options.FindOptions {
	opts := options.Find()

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}

	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}

	if q.skip > 0 {
		opts.SetSkip(q.skip)
	}

	if len(q.project) > 0 {
		opts.SetProjection(q.project)
	}

	return opts
}

// Reset 重置查詢構建器
func (q *QueryBuilder) Reset() *QueryBuilder {
	q.filter = bson.M{}
	q.sort = bson.D{}
	q.limit = 0
	q.skip = 0
	q.project = bson.M{}
	return q
}

// Clone 克隆查詢構建器
func (q *QueryBuilder) Clone() *QueryBuilder {
	newBuilder := NewQueryBuilder()

	// 深拷貝 filter
	for k, v := range q.filter {
		newBuilder.filter[k] = v
	}

	// 深拷貝 sort
	newBuilder.sort = append(newBuilder.sort, q.sort...)

	// 深拷貝 project
	for k, v := range q.project {
		newBuilder.project[k] = v
	}

	newBuilder.limit = q.limit
	newBuilder.skip = q.skip

	return newBuilder
}

// Count 返回符合條件的文檔數量
func (q *QueryBuilder) Count() *QueryBuilder {
	// 這個方法主要是為了語義完整性，實際計數需要在ODM中執行
	return q
}

// WhereNotExists 添加一個不存在條件
func (q *QueryBuilder) WhereNotExists(field string) *QueryBuilder {
	q.filter[field] = bson.M{"$exists": false}
	return q
}

// WhereIsNull 添加一個空值條件
func (q *QueryBuilder) WhereIsNull(field string) *QueryBuilder {
	q.filter[field] = nil
	return q
}

// WhereNotNull 添加一個非空值條件
func (q *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	q.filter[field] = bson.M{"$ne": nil}
	return q
}

// WhereBetween 添加一個範圍條件
func (q *QueryBuilder) WhereBetween(field string, min, max any) *QueryBuilder {
	q.filter[field] = bson.M{
		"$gte": min,
		"$lte": max,
	}
	return q
}

// WhereNotBetween 添加一個不在範圍內的條件
func (q *QueryBuilder) WhereNotBetween(field string, min, max any) *QueryBuilder {
	q.filter["$or"] = []bson.M{
		{field: bson.M{"$lt": min}},
		{field: bson.M{"$gt": max}},
	}
	return q
}

// WhereSize 添加一個陣列大小條件
func (q *QueryBuilder) WhereSize(field string, size int) *QueryBuilder {
	q.filter[field] = bson.M{"$size": size}
	return q
}

// WhereAll 添加一個包含所有元素的條件
func (q *QueryBuilder) WhereAll(field string, values any) *QueryBuilder {
	q.filter[field] = bson.M{"$all": values}
	return q
}

// WhereElemMatch 添加一個元素匹配條件
func (q *QueryBuilder) WhereElemMatch(field string, condition bson.M) *QueryBuilder {
	q.filter[field] = bson.M{"$elemMatch": condition}
	return q
}

// WhereRaw 添加原始查詢條件
func (q *QueryBuilder) WhereRaw(rawFilter bson.M) *QueryBuilder {
	for k, v := range rawFilter {
		q.filter[k] = v
	}
	return q
}

// Paginate 設置分頁參數
func (q *QueryBuilder) Paginate(page, pageSize int64) *QueryBuilder {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	q.skip = (page - 1) * pageSize
	q.limit = pageSize
	return q
}

// SortBy 通用排序方法
func (q *QueryBuilder) SortBy(field string, order int) *QueryBuilder {
	q.sort = append(q.sort, bson.E{Key: field, Value: order})
	return q
}

// GetQueryOptions 獲取查詢選項（用於ODM）
func (q *QueryBuilder) GetQueryOptions() *QueryOptions {
	opts := &QueryOptions{
		Sort:       q.sort,
		Projection: q.project,
	}

	if q.limit > 0 {
		opts.Limit = &q.limit
	}

	if q.skip > 0 {
		opts.Skip = &q.skip
	}

	return opts
}
