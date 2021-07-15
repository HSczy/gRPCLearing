package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/HSczy/gRPCLearning/route"
	"google.golang.org/grpc"
)
func runFirst(client pb.RouteGuideClient) {
	feature, err := client.GetFeature(context.Background(), &pb.Point{
		Latitude:  310020000,
		Longitude: 123440000,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(feature)
}

func runSecond(client pb.RouteGuideClient) {
	serverStream, err := client.ListFeatures(context.Background(), &pb.Rectangle{
		Lo: &pb.Point{
			Latitude:  310022514,
			Longitude: 123440410,
		},
		Hi: &pb.Point{
			Latitude:  151421410,
			Longitude: 151454241,
		}})
	if err != nil{
		log.Fatalln(err)
	}
	for {
		rec, err := serverStream.Recv()
		if err != nil {
			if err == io.EOF{
				break
			}
			log.Fatalln(err)
		}
		fmt.Println(rec)
	}
}

func runThird(client pb.RouteGuideClient) {
	points := []*pb.Point{
		{Latitude:  310020000, Longitude: 123440000},
		{Latitude:  310022514, Longitude: 123440410},
		{Latitude:  151421410, Longitude: 151454241},
	}
	clientStream, err := client.RecordRoute(context.Background())
	if err != nil{
		log.Fatalln(err)
	}
	for _, point := range points {
		if err := clientStream.Send(point);err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second)
	}
	summary, err := clientStream.CloseAndRecv()
	if err != nil{
		log.Fatalln(err)
	}
	fmt.Println(summary)
}
func readIntFromCommandLine(reader *bufio.Reader, target *int32) {
	_, err := fmt.Fscanf(reader, "%d\n", target)
	if err != nil {
		log.Fatalln("Cannot scan", err)
	}
}
func runFourth (client pb.RouteGuideClient) {
	stream, err := client.Recommend(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		feature, err2 := stream.Recv()
		if err2 != nil {
			log.Fatalln(err)
		}
		fmt.Println("Recommended: ",feature)
	}()
	newReader := bufio.NewReader(os.Stdin)
	for {
		request := pb.RecommendationRequest{Point: new(pb.Point)}
		var mode int32
		fmt.Print("Enter Recommendation Mode(0 for farthest ,1 for the nearest)")
		readIntFromCommandLine(newReader,&mode)
		fmt.Println("Enter Latitude:")
		readIntFromCommandLine(newReader,&request.Point.Latitude)
		fmt.Println("Enter Longitude:")
		readIntFromCommandLine(newReader,&request.Point.Longitude)
		request.Mode = pb.RecommendationMode_GetNearest
		if mode == 0{
			request.Mode = pb.RecommendationMode_GetFarthest
		}
		err := stream.Send(&request)
		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalln("client cannot dial grpc server")
	}
	defer  conn.Close()
	client := pb.NewRouteGuideClient(conn)
	//runFirst(client)
	//runSecond(client)
	//runThird(client)
	runFourth(client)
}
