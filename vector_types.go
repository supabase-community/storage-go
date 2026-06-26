package storage_go

type VectorDataType string

const (
	VectorDataTypeFloat32 VectorDataType = "f32"
	VectorDataTypeFloat64 VectorDataType = "f64"
)

type DistanceMetric string

const (
	DistanceMetricCosine    DistanceMetric = "cosine"
	DistanceMetricEuclidean DistanceMetric = "euclidean"
	DistanceMetricDot       DistanceMetric = "dot"
)

type VectorMetadata map[string]any
type VectorFilter map[string]any

type VectorBucket struct {
	Name         string `json:"name"`
	CreationTime int64  `json:"creationTime,omitempty"`
}

type VectorIndex struct {
	Name           string         `json:"name"`
	Dimension      int            `json:"dimension,omitempty"`
	DistanceMetric DistanceMetric `json:"distanceMetric,omitempty"`
	DataType       VectorDataType `json:"dataType,omitempty"`
	CreationTime   int64          `json:"creationTime,omitempty"`
}

type VectorData struct {
	ID       string         `json:"id"`
	Vector   []float64      `json:"vector,omitempty"`
	Metadata VectorMetadata `json:"metadata,omitempty"`
}

type VectorMatch struct {
	ID       string         `json:"id"`
	Score    float64        `json:"score,omitempty"`
	Vector   []float64      `json:"vector,omitempty"`
	Metadata VectorMetadata `json:"metadata,omitempty"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type VectorBucketResponse struct {
	VectorBucket VectorBucket `json:"vectorBucket"`
}

type ListVectorBucketsOptions struct {
	Prefix     string `json:"prefix,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

type ListVectorBucketsResponse struct {
	VectorBuckets []VectorBucket `json:"vectorBuckets"`
	NextToken     string         `json:"nextToken,omitempty"`
}

type CreateIndexOptions struct {
	Name           string         `json:"name"`
	Dimension      int            `json:"dimension,omitempty"`
	DistanceMetric DistanceMetric `json:"distanceMetric,omitempty"`
	DataType       VectorDataType `json:"dataType,omitempty"`
}

type VectorIndexResponse struct {
	VectorIndex VectorIndex `json:"vectorIndex"`
}

type ListIndexesOptions struct {
	Prefix     string `json:"prefix,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
	NextToken  string `json:"nextToken,omitempty"`
}

type ListIndexesResponse struct {
	VectorIndexes []VectorIndex `json:"vectorIndexes"`
	NextToken     string        `json:"nextToken,omitempty"`
}

type PutVectorsOptions struct {
	Vectors []VectorData `json:"vectors"`
}

type GetVectorsOptions struct {
	IDs             []string `json:"ids"`
	IncludeVectors  bool     `json:"includeVectors,omitempty"`
	IncludeMetadata bool     `json:"includeMetadata,omitempty"`
}

type GetVectorsResponse struct {
	Vectors []VectorData `json:"vectors"`
}

type ListVectorsOptions struct {
	Prefix          string `json:"prefix,omitempty"`
	MaxResults      int    `json:"maxResults,omitempty"`
	NextToken       string `json:"nextToken,omitempty"`
	IncludeVectors  bool   `json:"includeVectors,omitempty"`
	IncludeMetadata bool   `json:"includeMetadata,omitempty"`
}

type ListVectorsResponse struct {
	Vectors   []VectorData `json:"vectors"`
	NextToken string       `json:"nextToken,omitempty"`
}

type QueryVectorsOptions struct {
	Vector          []float64      `json:"vector"`
	TopK            int            `json:"topK,omitempty"`
	Filter          VectorFilter   `json:"filter,omitempty"`
	IncludeVectors  bool           `json:"includeVectors,omitempty"`
	IncludeMetadata bool           `json:"includeMetadata,omitempty"`
	Metadata        VectorMetadata `json:"metadata,omitempty"`
}

type QueryVectorsResponse struct {
	Success bool          `json:"success,omitempty"`
	Matches []VectorMatch `json:"matches,omitempty"`
}

type DeleteVectorsOptions struct {
	IDs    []string     `json:"ids,omitempty"`
	Filter VectorFilter `json:"filter,omitempty"`
	Prefix string       `json:"prefix,omitempty"`
}
