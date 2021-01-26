package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"blog-grpc-gateway/third_party/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	collection *mongo.Collection
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

// Unary
func (svr server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Printf("Create blog request %v\n", req)
	b := req.Blog
	log.Printf("blog passed is %v\n", b)

	if b == nil {
		log.Println("Empty request")
		return nil, status.Errorf(codes.Internal, "Empty request")
	}
	bi := blogItem{
		AuthorID: b.AuthorId,
		Title:    b.Title,
		Content:  b.Content,
	}

	res, err := svr.collection.InsertOne(ctx, bi)
	if err != nil {
		// Unary - SquareRoot method for CalculatorService server.go:133
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v", err))
	}
	log.Printf("CreateBlog res is %v", res)

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Internal Eror: Cannot convert to OID")
	}

	persistedBlog := blogpb.Blog{
		Id:       oid.Hex(),
		AuthorId: bi.AuthorID,
		Title:    bi.Title,
		Content:  bi.Content,
	}
	resp := blogpb.CreateBlogResponse{
		Blog: &persistedBlog,
	}
	return &resp, nil
}

// Unary- sends back not found grpc style error
func (s server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Printf("Read blog request %v\n", req)
	blogId := req.BlogId
	fmt.Printf("Going to Read Blog for blog Id: %v\n", blogId)

	// Basically, ObjectId is treated as the primary key within any MongoDB collection.
	// It is generated automatically whenever we create a new document within a new collection.
	// It is based on a 12-byte hexadecimal value as you can observe in the following syntax.
	// Sep 1, 2020
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse blog id")
	}

	// empty struct as decode will be used to populate db data
	var bi blogItem
	bi, err = findOneBlog(ctx, s.collection, bi, oid)
	if err != nil {
		log.Printf("ReadBlog Err is err %v\n", err)
		return nil, err
	}
	log.Printf("Readblog Found bi %v\n", bi)

	return &blogpb.ReadBlogResponse{
		Blog: blogItemToBlogPb(&bi),
	}, nil
}

func blogItemToBlogPb(item *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       item.ID.Hex(),
		AuthorId: item.AuthorID,
		Title:    item.Title,
		Content:  item.Content,
	}
}
func findOneBlog(ctx context.Context, coll *mongo.Collection, item blogItem, oid primitive.ObjectID) (blogItem, error) {
	result := coll.FindOne(ctx, bson.M{"_id": oid})
	// As in json
	if err := result.Decode(&item); err != nil {
		return blogItem{}, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err))
	}
	return item, nil
}

// Unary- before updating, does not found check, sends back not found grpc style error
func (s server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update blog request")
	blog := req.Blog
	fmt.Printf("Update blog request blog id passed is %v\n", blog.Id)
	oid, err := primitive.ObjectIDFromHex(blog.Id)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Cannot parse blog id")
	}

	// find the document for which the _id field matches id and set the email to "newemail@example.com"
	// specify the Upsert option to insert a new document if a document matching the filter isn't found
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": oid}

	fmt.Println("Trying to use blog item directly instead of bson.M since like map and update is interface{}")
	update := bson.M{
		"$set": blogItem{
			AuthorID: blog.AuthorId,
			Title:    blog.Title,
			Content:  blog.Content,
		},
	}
	// update := bson.M{
	// 	"$set": bson.M{"author_id": blog.AuthorId,
	// 		"title":   blog.Title,
	// 		"content": blog.Content,
	// 	},
	// }

	uResult, err := s.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot update for blog %v due to error %v\n", blog, err))
	}
	if uResult.MatchedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v", blog.Id))
	}
	if uResult.UpsertedCount != 0 {
		log.Printf("inserted a new document with ID %v\n", uResult.UpsertedID)
	}

	// read saved document
	var bi blogItem
	bi, err = findOneBlog(ctx, s.collection, bi, oid)
	if err != nil {
		return nil, err
	}

	log.Printf("Updateblog from reading Found bi %v\n", bi)
	ue := blogpb.UpdateBlogResponse{Blog: blogItemToBlogPb(&bi)}
	return &ue, nil
}

// Unary- sends back not found grpc style error
func (s server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete blog request")
	fmt.Printf("Delete blog request blog id passed is %v\n", req.BlogId)
	oid, err := primitive.ObjectIDFromHex(req.BlogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Cannot parse blog id")
	}
	filter := bson.M{"_id": oid}

	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot Parse Id")
	}

	if result.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Caould not fins blog in MongoBD %v", err)
	}

	return &blogpb.DeleteBlogResponse{BlogId: req.BlogId}, nil
}

// Server streaming
func (s server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) (err error) {
	fmt.Printf("ListBlog rpc request passed from client is %v\n", req)
	cur, err := s.collection.Find(context.Background(), primitive.D{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err))
	}
	defer func() {
		fmt.Println("defer effects; closing cur")
		err = cur.Close(context.Background())
		// err = errors.New("to see defer effects")
		if err != nil {
			// named return https://stackoverflow.com/questions/37248898/how-does-defer-and-named-return-value-work
			err = status.Errorf(
				codes.Internal,
				fmt.Sprintf("Unknown internal error: %v", err))
		}
	}()

	for cur.Next(context.Background()) {
		var bi blogItem
		if err := cur.Decode(&bi); err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data: %v", err))
		}
		writeBlogToStream(blogItemToBlogPb(&bi), stream)
	}
	// Several methods return a cursor, which can be u... doc.go:32
	// The doc.go is also helpful to hence always check
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}

	return nil
}

func writeBlogToStream(blog *blogpb.Blog, stream blogpb.BlogService_ListBlogServer) {
	resp := blogpb.ListBlogResponse{
		Blog: blog,
	}
	if err := stream.Send(&resp); err != nil {
		log.Fatalf("error in sending %v\n", err)
	}
}

func tlsOpts(opts []grpc.ServerOption) []grpc.ServerOption {
	certFile := "tlsdocker/server.crt"
	keyFile := "tlsdocker/server.pem"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("tls error failed loading certificates %v", err)
	}
	opts = []grpc.ServerOption{grpc.Creds(creds)}
	return opts
}

func main() {
	// we get filename and line number -- see the difference and so just by doing this we don't need stacktrace to great extent

	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	// growth@rsachdeva-mac-7 blog10.54svc % go run cmd/server/server.go
	// 	Server for blog service started
	// 2020/10/12 19:36:05 server.go:35: could not listen listen tcp 0.0.0.0:50051: bind: address already in use
	// exit status 1

	// with no log set
	// growth@rsachdeva-mac-7 blog10.54svc % go run cmd/server/server.go
	// 	Server for blog service started
	// 2020/10/12 19:36:18 could not listen listen tcp 0.0.0.0:50051: bind: address already in use
	// exit status 1

	//illuminatingdeposits we have
	// log.LstdFlags|log.Lmicroseconds|log.Llongfile

	// growth@rsachdeva-mac-7 blog10.54svc % go run cmd/server/server.go
	// 	Server for blog service started
	// 2020/10/12 19:45:53.433002 /Users/growth/drinnovations/mywork_jmd/gogrpc/gomasterclassgrpcudemy/grpc-go-course/mygocode/blog10.54svc/cmd/server/server.go:49: could not listen listen tcp 0.0.0.0:50051: bind: address already in use
	// exit status 1
	// log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)

	// here we have
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// MongoDB start up
	fmt.Println("Connecting to MongoDB")

	ctx, mt := connectMongoDB()

	fmt.Println("Server for blog service started")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("could not listen %v", err)
	}

	log.Println("Connecting to Blog collection..")
	svr := server{}
	svr.collection = mt.Database("myblogdb").Collection("blog")

	// since execution happens from root of project per the go.mod file
	// for evans turning it off
	// quick hack for demo
	tls := false
	fmt.Println("tls option is ", tls)
	var opts []grpc.ServerOption
	if tls {
		opts = tlsOpts(opts)
	}
	// https://golang.org/ref/spec#Passing_arguments_to_..._parameters
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, svr)

	// Register reflection service on gRPC server.
	reflection.Register(s)
	serveWithShutdown(s, lis, mt, ctx)
}
