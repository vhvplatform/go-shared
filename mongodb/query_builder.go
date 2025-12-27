package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QueryBuilder provides a fluent interface for building MongoDB queries
type QueryBuilder struct {
	filter bson.M
}

// NewQueryBuilder creates a new QueryBuilder
// Performance: Pre-allocate map with reasonable capacity
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		filter: make(bson.M, 8), // Pre-allocate for common query size
	}
}

// Where adds an equality condition
func (qb *QueryBuilder) Where(field string, value interface{}) *QueryBuilder {
	qb.filter[field] = value
	return qb
}

// WhereIn adds an $in condition
func (qb *QueryBuilder) WhereIn(field string, values interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$in": values}
	return qb
}

// WhereNotIn adds a $nin condition
func (qb *QueryBuilder) WhereNotIn(field string, values interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$nin": values}
	return qb
}

// WhereGreaterThan adds a $gt condition
func (qb *QueryBuilder) WhereGreaterThan(field string, value interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$gt": value}
	return qb
}

// WhereGreaterThanOrEqual adds a $gte condition
func (qb *QueryBuilder) WhereGreaterThanOrEqual(field string, value interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$gte": value}
	return qb
}

// WhereLessThan adds a $lt condition
func (qb *QueryBuilder) WhereLessThan(field string, value interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$lt": value}
	return qb
}

// WhereLessThanOrEqual adds a $lte condition
func (qb *QueryBuilder) WhereLessThanOrEqual(field string, value interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$lte": value}
	return qb
}

// WhereBetween adds a range condition ($gte and $lte)
func (qb *QueryBuilder) WhereBetween(field string, min, max interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{
		"$gte": min,
		"$lte": max,
	}
	return qb
}

// WhereRegex adds a regex pattern matching condition
func (qb *QueryBuilder) WhereRegex(field string, pattern string, options string) *QueryBuilder {
	qb.filter[field] = bson.M{
		"$regex":   pattern,
		"$options": options,
	}
	return qb
}

// WhereExists checks if a field exists
func (qb *QueryBuilder) WhereExists(field string, exists bool) *QueryBuilder {
	qb.filter[field] = bson.M{"$exists": exists}
	return qb
}

// WhereNull checks if a field is null
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	qb.filter[field] = nil
	return qb
}

// WhereNotNull checks if a field is not null
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	qb.filter[field] = bson.M{"$ne": nil}
	return qb
}

// WhereArrayContains checks if an array contains a value
func (qb *QueryBuilder) WhereArrayContains(field string, value interface{}) *QueryBuilder {
	qb.filter[field] = bson.M{"$elemMatch": bson.M{"$eq": value}}
	return qb
}

// WhereArraySize checks the size of an array
func (qb *QueryBuilder) WhereArraySize(field string, size int) *QueryBuilder {
	qb.filter[field] = bson.M{"$size": size}
	return qb
}

// Or adds an $or condition with multiple sub-conditions
func (qb *QueryBuilder) Or(conditions ...bson.M) *QueryBuilder {
	qb.filter["$or"] = conditions
	return qb
}

// And adds an $and condition with multiple sub-conditions
func (qb *QueryBuilder) And(conditions ...bson.M) *QueryBuilder {
	qb.filter["$and"] = conditions
	return qb
}

// WhereObjectID adds an ObjectID equality condition
// If the id string is invalid, the field will be set to match nothing (empty ObjectID)
func (qb *QueryBuilder) WhereObjectID(field string, id string) *QueryBuilder {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Use zero ObjectID which will not match any real document
		objectID = primitive.NilObjectID
	}
	qb.filter[field] = objectID
	return qb
}

// WhereDate adds a date equality condition (day precision)
func (qb *QueryBuilder) WhereDate(field string, date time.Time) *QueryBuilder {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	qb.filter[field] = bson.M{
		"$gte": startOfDay,
		"$lt":  endOfDay,
	}
	return qb
}

// WhereDateAfter adds a condition for dates after a given date
func (qb *QueryBuilder) WhereDateAfter(field string, date time.Time) *QueryBuilder {
	qb.filter[field] = bson.M{"$gt": date}
	return qb
}

// WhereDateBefore adds a condition for dates before a given date
func (qb *QueryBuilder) WhereDateBefore(field string, date time.Time) *QueryBuilder {
	qb.filter[field] = bson.M{"$lt": date}
	return qb
}

// WhereTextSearch adds a full-text search condition
func (qb *QueryBuilder) WhereTextSearch(text string) *QueryBuilder {
	qb.filter["$text"] = bson.M{"$search": text}
	return qb
}

// Build returns the final filter
func (qb *QueryBuilder) Build() bson.M {
	return qb.filter
}

// BuildWithTenant adds tenant_id to the filter and returns it
func (qb *QueryBuilder) BuildWithTenant(tenantID string) bson.M {
	qb.filter["tenant_id"] = tenantID
	return qb.filter
}

// Clone creates a copy of the QueryBuilder for reusability
// Performance: Pre-allocate with same capacity as source
func (qb *QueryBuilder) Clone() *QueryBuilder {
	newFilter := make(bson.M, len(qb.filter))
	for k, v := range qb.filter {
		newFilter[k] = v
	}
	return &QueryBuilder{
		filter: newFilter,
	}
}

// Reset clears all conditions
// Performance: Reuse underlying map storage
func (qb *QueryBuilder) Reset() *QueryBuilder {
	// Clear map but keep allocated capacity
	for k := range qb.filter {
		delete(qb.filter, k)
	}
	return qb
}

// AggregationBuilder provides a fluent interface for building aggregation pipelines
type AggregationBuilder struct {
	pipeline []bson.M
}

// NewAggregationBuilder creates a new AggregationBuilder
// Performance: Pre-allocate slice with reasonable capacity
func NewAggregationBuilder() *AggregationBuilder {
	return &AggregationBuilder{
		pipeline: make([]bson.M, 0, 8), // Pre-allocate for common pipeline size
	}
}

// Match adds a $match stage
func (ab *AggregationBuilder) Match(filter bson.M) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$match": filter})
	return ab
}

// Group adds a $group stage
func (ab *AggregationBuilder) Group(group bson.M) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$group": group})
	return ab
}

// Sort adds a $sort stage
func (ab *AggregationBuilder) Sort(sort bson.D) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$sort": sort})
	return ab
}

// Limit adds a $limit stage
func (ab *AggregationBuilder) Limit(limit int64) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$limit": limit})
	return ab
}

// Skip adds a $skip stage
func (ab *AggregationBuilder) Skip(skip int64) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$skip": skip})
	return ab
}

// Project adds a $project stage
func (ab *AggregationBuilder) Project(projection bson.M) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$project": projection})
	return ab
}

// Lookup adds a $lookup stage for joining collections
func (ab *AggregationBuilder) Lookup(from, localField, foreignField, as string) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return ab
}

// LookupWithPipeline adds a $lookup stage with a custom pipeline for complex joins
func (ab *AggregationBuilder) LookupWithPipeline(from, as string, let bson.M, pipeline []bson.M) *AggregationBuilder {
	lookupStage := bson.M{
		"from": from,
		"as":   as,
	}
	if let != nil {
		lookupStage["let"] = let
	}
	if pipeline != nil {
		lookupStage["pipeline"] = pipeline
	}
	ab.pipeline = append(ab.pipeline, bson.M{"$lookup": lookupStage})
	return ab
}

// PopulateField adds lookup and unwind stages to populate a foreign key field
// This is a convenience method that combines $lookup and $unwind
func (ab *AggregationBuilder) PopulateField(from, localField, foreignField, as string, preserveNull bool) *AggregationBuilder {
	// Add lookup
	ab.Lookup(from, localField, foreignField, as)

	// Unwind the result array, optionally preserving null/empty
	if preserveNull {
		ab.UnwindWithOptions("$"+as, true)
	} else {
		ab.Unwind("$" + as)
	}

	return ab
}

// PopulateConfig defines configuration for populating a foreign key field
type PopulateConfig struct {
	From         string   // Source collection name
	LocalField   string   // Field in current document
	ForeignField string   // Field in foreign collection (usually "_id")
	As           string   // Output field name
	PreserveNull bool     // Whether to preserve documents when foreign key is null
	Fields       []string // Specific fields to include from foreign collection (nil = all fields)
}

// PopulateMultiple adds lookup stages for multiple foreign keys with field selection
func (ab *AggregationBuilder) PopulateMultiple(configs []PopulateConfig) *AggregationBuilder {
	for _, config := range configs {
		if config.Fields != nil && len(config.Fields) > 0 {
			// Use pipeline to select specific fields
			let := bson.M{
				"localField": "$" + config.LocalField,
			}

			// Build project stage to select fields
			projectStage := bson.M{"_id": 1}
			for _, field := range config.Fields {
				projectStage[field] = 1
			}

			pipeline := []bson.M{
				{"$match": bson.M{"$expr": bson.M{"$eq": []interface{}{"$" + config.ForeignField, "$$localField"}}}},
				{"$project": projectStage},
			}

			ab.LookupWithPipeline(config.From, config.As, let, pipeline)
		} else {
			// Simple lookup without field selection
			ab.Lookup(config.From, config.LocalField, config.ForeignField, config.As)
		}

		// Unwind if needed
		if config.PreserveNull {
			ab.UnwindWithOptions("$"+config.As, true)
		} else {
			ab.Unwind("$" + config.As)
		}
	}
	return ab
}

// Unwind adds an $unwind stage
func (ab *AggregationBuilder) Unwind(path string) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$unwind": path})
	return ab
}

// UnwindWithOptions adds an $unwind stage with options
func (ab *AggregationBuilder) UnwindWithOptions(path string, preserveNullAndEmptyArrays bool) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{
		"$unwind": bson.M{
			"path":                       path,
			"preserveNullAndEmptyArrays": preserveNullAndEmptyArrays,
		},
	})
	return ab
}

// Facet adds a $facet stage for multi-faceted aggregation
// Each facet runs a separate sub-pipeline on the same set of input documents
func (ab *AggregationBuilder) Facet(facets map[string][]bson.M) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$facet": facets})
	return ab
}

// Bucket adds a $bucket stage for grouping documents into buckets
func (ab *AggregationBuilder) Bucket(groupBy interface{}, boundaries []interface{}, defaultBucket string, output bson.M) *AggregationBuilder {
	bucketStage := bson.M{
		"groupBy":    groupBy,
		"boundaries": boundaries,
	}
	if defaultBucket != "" {
		bucketStage["default"] = defaultBucket
	}
	if output != nil {
		bucketStage["output"] = output
	}
	ab.pipeline = append(ab.pipeline, bson.M{"$bucket": bucketStage})
	return ab
}

// BucketAuto adds a $bucketAuto stage for automatic bucketing
func (ab *AggregationBuilder) BucketAuto(groupBy interface{}, buckets int, output bson.M, granularity string) *AggregationBuilder {
	bucketAutoStage := bson.M{
		"groupBy": groupBy,
		"buckets": buckets,
	}
	if output != nil {
		bucketAutoStage["output"] = output
	}
	if granularity != "" {
		bucketAutoStage["granularity"] = granularity
	}
	ab.pipeline = append(ab.pipeline, bson.M{"$bucketAuto": bucketAutoStage})
	return ab
}

// AddFields adds a $addFields stage to add new fields to documents
func (ab *AggregationBuilder) AddFields(fields bson.M) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$addFields": fields})
	return ab
}

// ReplaceRoot adds a $replaceRoot stage to replace the document root
func (ab *AggregationBuilder) ReplaceRoot(newRoot string) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{
		"$replaceRoot": bson.M{"newRoot": newRoot},
	})
	return ab
}

// Sample adds a $sample stage to randomly select documents
func (ab *AggregationBuilder) Sample(size int64) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{
		"$sample": bson.M{"size": size},
	})
	return ab
}

// Count adds a $count stage to count documents
func (ab *AggregationBuilder) Count(field string) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$count": field})
	return ab
}

// SortByCount adds a $sortByCount stage to group and count by expression
func (ab *AggregationBuilder) SortByCount(expression interface{}) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$sortByCount": expression})
	return ab
}

// Build returns the aggregation pipeline
func (ab *AggregationBuilder) Build() []bson.M {
	return ab.pipeline
}

// Clone creates a copy of the AggregationBuilder
func (ab *AggregationBuilder) Clone() *AggregationBuilder {
	newPipeline := make([]bson.M, len(ab.pipeline))
	copy(newPipeline, ab.pipeline)
	return &AggregationBuilder{
		pipeline: newPipeline,
	}
}

// Reset clears all stages
func (ab *AggregationBuilder) Reset() *AggregationBuilder {
	ab.pipeline = []bson.M{}
	return ab
}

// StatsHelper provides quick statistics aggregation helpers
type StatsHelper struct{}

// NewStatsHelper creates a new StatsHelper
func NewStatsHelper() *StatsHelper {
	return &StatsHelper{}
}

// CountByField creates a pipeline to count documents grouped by a field
func (sh *StatsHelper) CountByField(field string) []bson.M {
	return []bson.M{
		{"$group": bson.M{
			"_id":   "$" + field,
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.D{{Key: "count", Value: -1}}},
	}
}

// SumByField creates a pipeline to sum a numeric field grouped by another field
func (sh *StatsHelper) SumByField(groupField, sumField string) []bson.M {
	return []bson.M{
		{"$group": bson.M{
			"_id":   "$" + groupField,
			"total": bson.M{"$sum": "$" + sumField},
		}},
		{"$sort": bson.D{{Key: "total", Value: -1}}},
	}
}

// AverageByField creates a pipeline to calculate average of a numeric field grouped by another field
func (sh *StatsHelper) AverageByField(groupField, avgField string) []bson.M {
	return []bson.M{
		{"$group": bson.M{
			"_id":     "$" + groupField,
			"average": bson.M{"$avg": "$" + avgField},
			"count":   bson.M{"$sum": 1},
		}},
		{"$sort": bson.D{{Key: "average", Value: -1}}},
	}
}

// MinMaxByField creates a pipeline to find min and max of a field grouped by another field
func (sh *StatsHelper) MinMaxByField(groupField, valueField string) []bson.M {
	return []bson.M{
		{"$group": bson.M{
			"_id": "$" + groupField,
			"min": bson.M{"$min": "$" + valueField},
			"max": bson.M{"$max": "$" + valueField},
		}},
		{"$sort": bson.D{{Key: "_id", Value: 1}}},
	}
}

// StatsForField creates a comprehensive statistics pipeline for a numeric field
func (sh *StatsHelper) StatsForField(field string) []bson.M {
	return []bson.M{
		{"$group": bson.M{
			"_id":   nil,
			"count": bson.M{"$sum": 1},
			"sum":   bson.M{"$sum": "$" + field},
			"avg":   bson.M{"$avg": "$" + field},
			"min":   bson.M{"$min": "$" + field},
			"max":   bson.M{"$max": "$" + field},
		}},
	}
}

// TopN creates a pipeline to get top N documents by a field
func (sh *StatsHelper) TopN(field string, n int) []bson.M {
	return []bson.M{
		{"$sort": bson.D{{Key: field, Value: -1}}},
		{"$limit": n},
	}
}

// BottomN creates a pipeline to get bottom N documents by a field
func (sh *StatsHelper) BottomN(field string, n int) []bson.M {
	return []bson.M{
		{"$sort": bson.D{{Key: field, Value: 1}}},
		{"$limit": n},
	}
}

// DateRangeStats creates a pipeline to group and count documents by date ranges
func (sh *StatsHelper) DateRangeStats(dateField string, groupBy string) []bson.M {
	var groupExpr interface{}
	switch groupBy {
	case "year":
		groupExpr = bson.M{"$year": "$" + dateField}
	case "month":
		groupExpr = bson.M{
			"year":  bson.M{"$year": "$" + dateField},
			"month": bson.M{"$month": "$" + dateField},
		}
	case "day":
		groupExpr = bson.M{
			"year":  bson.M{"$year": "$" + dateField},
			"month": bson.M{"$month": "$" + dateField},
			"day":   bson.M{"$dayOfMonth": "$" + dateField},
		}
	default:
		groupExpr = bson.M{"$dateToString": bson.M{
			"format": "%Y-%m-%d",
			"date":   "$" + dateField,
		}}
	}

	return []bson.M{
		{"$group": bson.M{
			"_id":   groupExpr,
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.D{{Key: "_id", Value: 1}}},
	}
}

// PercentileStats creates a pipeline to calculate percentiles for a numeric field
func (sh *StatsHelper) PercentileStats(field string, percentiles []float64) []bson.M {
	return []bson.M{
		{"$sort": bson.D{{Key: field, Value: 1}}},
		{"$group": bson.M{
			"_id":    nil,
			"values": bson.M{"$push": "$" + field},
			"count":  bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"count": 1,
			"percentiles": bson.M{
				"$map": bson.M{
					"input": percentiles,
					"as":    "p",
					"in": bson.M{
						"percentile": "$$p",
						"value": bson.M{
							"$arrayElemAt": []interface{}{
								"$values",
								bson.M{
									"$floor": bson.M{
										"$multiply": []interface{}{
											bson.M{"$divide": []interface{}{"$$p", 100}},
											bson.M{"$subtract": []interface{}{"$count", 1}},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
	}
}

// PopulateHelper provides utilities for populating foreign key relationships
type PopulateHelper struct{}

// NewPopulateHelper creates a new PopulateHelper
func NewPopulateHelper() *PopulateHelper {
	return &PopulateHelper{}
}

// LookupConfig defines configuration for a single lookup operation
type LookupConfig struct {
	From         string            // Source collection name
	LocalField   string            // Field in current document containing foreign key
	ForeignField string            // Field in foreign collection to match (usually "_id")
	As           string            // Output field name in result
	PreserveNull bool              // Keep documents when foreign key is null/missing
	SelectFields []string          // Specific fields to select from foreign collection
	RenameFields map[string]string // Map of original field names to new names
}

// BuildPopulatePipeline creates an aggregation pipeline to populate foreign keys
// This can handle multiple lookups with field selection and renaming
func (ph *PopulateHelper) BuildPopulatePipeline(configs []LookupConfig) []bson.M {
	pipeline := []bson.M{}

	for _, config := range configs {
		// Build the lookup stage
		if config.SelectFields != nil || config.RenameFields != nil {
			// Use pipeline lookup for field selection/renaming
			let := bson.M{
				"localField": "$" + config.LocalField,
			}

			subPipeline := []bson.M{
				{"$match": bson.M{"$expr": bson.M{"$eq": []interface{}{"$" + config.ForeignField, "$$localField"}}}},
			}

			// Add project stage for field selection and renaming
			if config.SelectFields != nil || config.RenameFields != nil {
				projectStage := bson.M{}

				if config.SelectFields != nil {
					for _, field := range config.SelectFields {
						// Check if this field should be renamed
						if config.RenameFields != nil {
							if newName, exists := config.RenameFields[field]; exists {
								projectStage[newName] = "$" + field
							} else {
								projectStage[field] = 1
							}
						} else {
							projectStage[field] = 1
						}
					}
				} else if config.RenameFields != nil {
					// Only renaming, include all fields but rename specified ones
					for oldName, newName := range config.RenameFields {
						projectStage[newName] = "$" + oldName
					}
				}

				if len(projectStage) > 0 {
					subPipeline = append(subPipeline, bson.M{"$project": projectStage})
				}
			}

			lookupStage := bson.M{
				"from":     config.From,
				"let":      let,
				"pipeline": subPipeline,
				"as":       config.As,
			}
			pipeline = append(pipeline, bson.M{"$lookup": lookupStage})
		} else {
			// Simple lookup without field manipulation
			pipeline = append(pipeline, bson.M{
				"$lookup": bson.M{
					"from":         config.From,
					"localField":   config.LocalField,
					"foreignField": config.ForeignField,
					"as":           config.As,
				},
			})
		}

		// Add unwind stage
		if config.PreserveNull {
			pipeline = append(pipeline, bson.M{
				"$unwind": bson.M{
					"path":                       "$" + config.As,
					"preserveNullAndEmptyArrays": true,
				},
			})
		} else {
			pipeline = append(pipeline, bson.M{
				"$unwind": "$" + config.As,
			})
		}
	}

	return pipeline
}

// PopulateSingle creates a simple lookup pipeline for a single foreign key
func (ph *PopulateHelper) PopulateSingle(from, localField, foreignField, as string, preserveNull bool) []bson.M {
	return ph.BuildPopulatePipeline([]LookupConfig{
		{
			From:         from,
			LocalField:   localField,
			ForeignField: foreignField,
			As:           as,
			PreserveNull: preserveNull,
		},
	})
}

// PopulateWithFields creates a lookup pipeline with field selection
func (ph *PopulateHelper) PopulateWithFields(from, localField, foreignField, as string, fields []string, preserveNull bool) []bson.M {
	return ph.BuildPopulatePipeline([]LookupConfig{
		{
			From:         from,
			LocalField:   localField,
			ForeignField: foreignField,
			As:           as,
			SelectFields: fields,
			PreserveNull: preserveNull,
		},
	})
}

// PopulateWithRename creates a lookup pipeline with field renaming
func (ph *PopulateHelper) PopulateWithRename(from, localField, foreignField, as string, renameMap map[string]string, preserveNull bool) []bson.M {
	return ph.BuildPopulatePipeline([]LookupConfig{
		{
			From:         from,
			LocalField:   localField,
			ForeignField: foreignField,
			As:           as,
			RenameFields: renameMap,
			PreserveNull: preserveNull,
		},
	})
}
