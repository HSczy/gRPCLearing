package main

import (
	"context"
	"io"
	"log"
	"math"
	"net"
	"time"

	pb "github.com/HSczy/gRPCLearning/route"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type routeGuiderServer struct {
	pb.UnimplementedRouteGuideServer
	db []*pb.Feature
}

func (s *routeGuiderServer) GetFeature(ctx context.Context,point *pb.Point) (*pb.Feature, error) {
	for _,feature := range s.db{
		if proto.Equal(feature.Location,point){
			return feature, nil
		}
	}
	return nil, nil
}

func inRange(point *pb.Point,rect *pb.Rectangle)bool {
	left := math.Min(float64(rect.Lo.Longitude),float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude),float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude),float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude),float64(rect.Hi.Latitude))

	if float64(point.Longitude) >= left &&
		float64(point.Longitude) <= right &&
		float64(point.Latitude) >= bottom &&
		float64(point.Latitude) <= top{
		return true
	}
	return false
}


func (s *routeGuiderServer) ListFeatures(rectangle *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	for _,feature := range s.db{
		time.Sleep(time.Second*3)
		if inRange(feature.Location,rectangle){
			// 将找到的内容传递给流请求端去处理
			if err := stream.Send(feature);err != nil {
				return err
			}
		}
	}
	return nil
}

func toRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

// calcDistance calculates the distance between two points using the "haversine" formula.
// The formula is based on http://mathforum.org/library/drmath/view/51879.html.
func calcDistance(p1 *pb.Point, p2 *pb.Point) int32 {
	const CordFactor float64 = 1e7
	const R = float64(6371000) // earth radius in metres
	lat1 := toRadians(float64(p1.Latitude) / CordFactor)
	lat2 := toRadians(float64(p2.Latitude) / CordFactor)
	lng1 := toRadians(float64(p1.Longitude) / CordFactor)
	lng2 := toRadians(float64(p2.Longitude) / CordFactor)
	dlat := lat2 - lat1
	dlng := lng2 - lng1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return int32(distance)
}

func (s *routeGuiderServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	startTime := time.Now()
	var pointCount,distance int32
	var prevPoint *pb.Point

	for {
		point, err := stream.Recv()
		if err != nil {
			if err == io.EOF{
				endTime := time.Now()
				return stream.SendAndClose(&pb.RouteSummary{
					PointCount:  pointCount,
					Distance:    distance,
					ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
				})
			}
			log.Fatalln(err)
		}
		pointCount ++
		if prevPoint != nil {
			distance += calcDistance(prevPoint,point)
		}
		prevPoint = point
	}

	return nil
}
func (s *routeGuiderServer) recommendOnce(request *pb.RecommendationRequest) (*pb.Feature, error) {
	var nearest, farthest *pb.Feature
	var nearestDistance, farthestDistance int32

	for _, feature := range s.db {
		distance := calcDistance(feature.Location, request.Point)
		if nearest == nil || distance < nearestDistance {
			nearestDistance = distance
			nearest = feature
		}
		if farthest == nil || distance > farthestDistance {
			farthestDistance = distance
			farthest = feature
		}
	}
	if request.Mode == pb.RecommendationMode_GetFarthest {
		return farthest, nil
	} else {
		return nearest, nil
	}
}

func (s *routeGuiderServer) Recommend(stream pb.RouteGuide_RecommendServer) error {
	for {
		request, err := stream.Recv()
		if err != nil {
			if err == io.EOF{
				return nil
			}
			return err
		}
		recommendOnce, err := s.recommendOnce(request)
		if err != nil {
			return err
		}
		return stream.Send(recommendOnce)
	}
}

func newServer() *routeGuiderServer {
	return &routeGuiderServer{
		db: []*pb.Feature{
			{Name: "测试地点1",
				Location: &pb.Point{
					Latitude:  310020000,
					Longitude: 123440000}},
			{Name: "测试地点2",
				Location: &pb.Point{
					Latitude:  310022514,
					Longitude: 123440410}},
			{Name: "测试地点3",
				Location: &pb.Point{
					Latitude:  151421410,
					Longitude: 151454241}},
		},
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		log.Fatalln("cannot create a listener at the address")
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	log.Fatalln(grpcServer.Serve(lis))
}
