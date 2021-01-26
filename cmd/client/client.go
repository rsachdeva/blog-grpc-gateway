package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"blog-grpc-gateway/third_party/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func tlsOpts() grpc.DialOption {
	certFile := "tlsdocker/ca.crt"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatalf("loading certificate error is %v", err)
	}
	opts := grpc.WithTransportCredentials(creds)
	return opts
}

func main() {
	fmt.Println("Blog client")

	// must be in sync with gRPC server
	// quick hack for demo
	tls := false
	opts := grpc.WithInsecure()
	if tls {
		opts = tlsOpts()
	}
	cc, err := grpc.Dial("localhost:50051", opts)
	log.Println("connected to server successfully")
	if err != nil {
		log.Fatalf("err is %v", err)
	}
	defer cc.Close()

	fmt.Println("Creating Blog...")
	c := blogpb.NewBlogServiceClient(cc)
	crBl, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "Rohit100",
			Title:    "Go gRPC REST Gateway Rocks!",
			Content:  "Go Rock everything -- gRPC REST Gateway  JMD JMD JMD",
		},
	})

	if err != nil {
		log.Fatalf("error returned is %v", err)
	}

	fmt.Printf("Blog created:%v\n", crBl.Blog)

	fmt.Println("Reading Blog...")
	// blogIdHexFound := "5f872bb1850c17afef57566f"
	// blogIdHexNotFound := "5f872bb1850c10afef57566f"
	// blogIdHexCannotParse := "5f872bb1850c17afef57566g"
	blogIdsToRead := []string{
		// blogIdHexFound,
		// blogIdHexNotFound,
		// blogIdHexCannotParse,
		crBl.Blog.Id,
	}
	for _, bId := range blogIdsToRead {
		reBl, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: bId})
		if err != nil {
			log.Printf("ReadBlog error %v for blog id passed %v\n", err, bId)
			continue
		}
		log.Printf("ReadBlog found is %v\n", reBl)
	}

	// uId := crBl.Blog.Id
	uId := "5f8758cbc3bb89c067f9c365"
	fmt.Printf("Updating Blog for blog Id... %v\n", uId)
	upBl := &blogpb.Blog{
		Id:       uId,
		AuthorId: "ByIdDeepikaFinest!JMD",
		Title:    "ByIdMy First Blog (edited)",
		Content:  "ByIdContent of the first blog, with some awesome additions!",
	}
	updateRes, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: upBl})
	if updateErr != nil {
		fmt.Printf("Error happened while updating: %v \n", updateErr)
	}
	fmt.Printf("Blog was updated: %v\n", updateRes)

	// uId := crBl.Blog.Id
	dId := "5f877f835e24d5f1c444c6b9"
	fmt.Printf("Deleting Blog for blog Id... %v\n", uId)
	delResp, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: dId})
	if err != nil {
		fmt.Printf("Error happened while deleting: %v \n", err)
	}
	fmt.Printf("Blog was deleted: %v \n", delResp)

	fmt.Printf("Streaming to list all blogs...")
	stream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("err is %v \v", err)
	}
	fmt.Println("Result Streamed Blogs : ")
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("err is", err)
		}
		fmt.Printf("%v  \n", resp.Blog)
		// time.Sleep(time.Second)
	}

}
