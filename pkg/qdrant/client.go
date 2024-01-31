package qdrant

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"strconv"
	"strings"

	pb "github.com/qdrant/go-client/qdrant"
	"github.com/webws/go-moda/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ErrNotFound      = "Not found"
	ErrAlreadyExists = "already exists"
)

type QdrantClient struct {
	grpcConn        *grpc.ClientConn
	collection      string
	size            uint64
	memmapThreshold uint64
	hnswOndisk      bool
	hnswM           uint64
	hnswEFConstruct uint64
}

func (qc *QdrantClient) Close() {
	qc.grpcConn.Close()
}

func (qc *QdrantClient) Collection() pb.CollectionsClient {
	return pb.NewCollectionsClient(qc.grpcConn)
}

func NewQdrantClient(qdrantAddr, collection string, size, memmapThreshold uint64, hnswOndisk bool, hnswM, hnswEFConstruct uint64) *QdrantClient {
	conn, err := grpc.Dial(qdrantAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalw("did not connect", "err", err)
	}
	return &QdrantClient{
		grpcConn:        conn,
		collection:      collection,
		size:            size,
		memmapThreshold: memmapThreshold,
		hnswOndisk:      hnswOndisk,
		hnswM:           hnswM,
		hnswEFConstruct: hnswEFConstruct,
	}
}

func toPayload(payload map[string]string) map[string]*pb.Value {
	ret := make(map[string]*pb.Value)
	for k, v := range payload {
		ret[k] = &pb.Value{Kind: &pb.Value_StringValue{StringValue: v}}
	}
	return ret
}

func (qc *QdrantClient) GetCollectionName() string {
	return qc.collection
}

func (qc *QdrantClient) GetVectorSize() uint64 {
	return qc.size
}

func (qc *QdrantClient) DeleteCollection(name string) error {
	cc := pb.NewCollectionsClient(qc.grpcConn)
	_, err := cc.Delete(context.TODO(), &pb.DeleteCollection{
		CollectionName: name,
	})
	return err
}

func (qc *QdrantClient) CreateCollection(name string, size uint64) error {
	cc := pb.NewCollectionsClient(qc.grpcConn)

	quantizationAlwaysRam := true
	//quantizationQuantile := float32(0.99)
	req := &pb.CreateCollection{
		CollectionName: name,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     size,
					Distance: pb.Distance_Cosine,
				},
			},
		},
		OptimizersConfig: &pb.OptimizersConfigDiff{
			MemmapThreshold: &qc.memmapThreshold,
		},
		HnswConfig: &pb.HnswConfigDiff{
			OnDisk:      &qc.hnswOndisk,
			M:           &qc.hnswM,
			EfConstruct: &qc.hnswEFConstruct,
		},
		QuantizationConfig: &pb.QuantizationConfig{
			Quantization: &pb.QuantizationConfig_Binary{
				Binary: &pb.BinaryQuantization{
					AlwaysRam: &quantizationAlwaysRam,
				},
			},
		},
	}
	_, err := cc.Create(context.Background(), req)
	if err != nil && strings.Contains(err.Error(), ErrAlreadyExists) {
		return nil
	}
	if err != nil {
		logger.Errorw("CreateCollection", "err", err)
		return err
	}
	return nil
}

func (qc *QdrantClient) CreatePoints(points []*pb.PointStruct) error {
	pc := pb.NewPointsClient(qc.grpcConn)

	wait := true
	pointsReq := pb.UpsertPoints{
		CollectionName: qc.collection,
		Points:         points,
		Wait:           &wait,
	}

	_, err := pc.Upsert(context.TODO(), &pointsReq)
	if err != nil {
		logger.Errorw("CreatePoints fail", "err", err)
		return err
	}
	return nil
}

func (qc *QdrantClient) CreatePoint(uuid string, collection string, vector []float32, payload map[string]string) error {
	point := &pb.PointStruct{}
	point.Id = &pb.PointId{
		PointIdOptions: &pb.PointId_Uuid{
			Uuid: uuid,
		},
	}
	point.Vectors = &pb.Vectors{
		VectorsOptions: &pb.Vectors_Vector{
			Vector: &pb.Vector{
				Data: vector,
			},
		},
	}
	point.Payload = toPayload(payload)

	pc := pb.NewPointsClient(qc.grpcConn)

	wait := true
	points := pb.UpsertPoints{
		CollectionName: collection,
		Points:         []*pb.PointStruct{point},
		Wait:           &wait,
	}

	_, err := pc.Upsert(context.TODO(), &points)
	if err != nil {
		return err
	}
	return nil
}

func (qc *QdrantClient) Search(ctx context.Context, vector []float32) ([]*pb.ScoredPoint, error) {
	sc := pb.NewPointsClient(qc.grpcConn)

	var strArr []string
	for _, v := range vector {
		strArr = append(strArr, strconv.FormatFloat(float64(v), 'f', -1, 64))
	}

	log.Println("vector: ", strings.Join(strArr, ", "))

	offset := uint64(0)
	searchResponse, err := sc.Search(ctx, &pb.SearchPoints{
		CollectionName: qc.collection,
		Vector:         vector,
		Limit:          3,
		Offset:         &offset,
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Include{
				Include: &pb.PayloadIncludeSelector{
					Fields: []string{"combined"},
				},
			},
		},
	})
	if err != nil && strings.Contains(err.Error(), ErrNotFound) {
		if err := qc.CreateCollection(qc.collection, qc.size); err != nil {
			logger.Errorw("search vector failed", "err", err)
			return nil, err
		}
		return qc.Search(ctx, vector)
	}

	if err != nil {
		return nil, err
	}

	if len(searchResponse.Result) == 0 {
		return nil, nil
	}

	return searchResponse.Result, nil
}

func (qc *QdrantClient) Scroll(ctx context.Context, query string) ([]*pb.RetrievedPoint, error) {
	sc := pb.NewPointsClient(qc.grpcConn)

	var shouldMatches []*pb.Condition
	for _, v := range strings.Split(query, " ") {
		shouldMatches = append(shouldMatches, &pb.Condition{
			ConditionOneOf: &pb.Condition_Field{
				Field: &pb.FieldCondition{
					Key: "combined",
					Match: &pb.Match{
						MatchValue: &pb.Match_Text{
							Text: v,
						},
					},
				},
			},
		})
	}

	limitScroll := uint32(2)
	scrollResponse, err := sc.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: qc.collection,
		Limit:          &limitScroll,
		Filter: &pb.Filter{
			Should: shouldMatches,
		},
	})
	if err != nil && strings.Contains(err.Error(), ErrNotFound) {
		if err := qc.CreateCollection(qc.collection, qc.size); err != nil {
			logger.Errorw("scroll failed", "err", err)
			return nil, err
		}
		return qc.Scroll(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	if len(scrollResponse.Result) == 0 {
		return nil, nil
	}

	return scrollResponse.Result, nil
}

func (qc *QdrantClient) MultiScroll(ctx context.Context, query string) ([]*pb.RetrievedPoint, error) {
	sc := pb.NewPointsClient(qc.grpcConn)

	limit := uint32(1)
	var scroll []*pb.RetrievedPoint
	group, _ := errgroup.WithContext(ctx)

	for _, v := range strings.Split(query, " ") {
		v := v
		group.Go(func() error {
			mustMatch := []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "raw",
							Match: &pb.Match{
								MatchValue: &pb.Match_Text{
									Text: v,
								},
							},
						},
					},
				},
			}

			response, err := sc.Scroll(ctx, &pb.ScrollPoints{
				CollectionName: qc.collection,
				Limit:          &limit,
				Filter: &pb.Filter{
					Must: mustMatch,
				},
			})
			if err != nil {
				return err
			}

			if len(response.Result) > 0 {
				scroll = append(scroll, response.Result[0])
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	var result []*pb.RetrievedPoint
	existScroll := make(map[string]bool)
	for _, v := range scroll {
		if _, ok := existScroll[v.Id.GetUuid()]; ok {
			continue
		}

		existScroll[v.Id.GetUuid()] = true
		result = append(result, v)
	}

	return result, nil
}
