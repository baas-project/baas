package machines

import (
	"context"
)

// ipaddr must be without port
func protoClient(ctx context.Context, ipaddr string) {
	//
	//var conn *grpc.ClientConn
	//
	//for {
	//	log.Printf("Trying to connect to node with ip address \"%v\":50051\n", ipaddr)
	//
	//	var err error
	//	conn, err = grpc.Dial(ipaddr + ":50051", grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
	//		Backoff:           backoff.DefaultConfig,
	//		MinConnectTimeout: 5 * time.Second,
	//	}))
	//	if err == nil {
	//		break
	//	}
	//
	//	log.Printf("error: %v\n", err)
	//	time.Sleep(1 * time.Second)
	//}
	//
	//client := pb.NewGreeterClient(conn)
	//
	//sendctx, cancel := context.WithDeadline(ctx, time.Now().Add(10 * time.Second))
	//defer cancel()
	//
	//var reply *pb.HelloReply
	//
	//for {
	//	var err error
	//	reply, err = client.SayHello(sendctx, &pb.HelloRequest{
	//		Name: "Jonathan",
	//	})
	//	if err == nil {
	//		break
	//	}
	//
	//	log.Printf("Another error occured: %v\n", err)
	//	time.Sleep(1 * time.Second)
	//}
	//
	//
	//fmt.Printf("reply: %v", reply)
	//
	//err := conn.Close()
	//if err != nil {
	//	fmt.Errorf("%v", err)
	//}
}
