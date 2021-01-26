# Plugin style is the right way
#protoc -I=simple --go_out=plugins=grpc:go simple/simple.proto

# protoc --go_out=plugins=grpc:greetpb greetpb/greetemptyservice.proto
#protoc -I=greetpb --go_out=plugins=grpc:greetpb greetpb/greet.proto

# grpc gateway
# go get -u github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
# go get -u github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
# Copied from
# ~/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.0.0/third_party/googleapis
# the google folder
# to
# ~/drinnovations/mywork_jmd/gogrpc/gomasterclassgrpcudemy/blog-grpc-gateway/api/blog/google
# protoc -I=api/blog --go_out=plugins=grpc:third_party/blogpb --grpc-gateway_out=logtostderr=true:third_party/blogpb blog.proto

protoc -I=api/blog --openapiv2_out third_party/blogopenapi --openapiv2_opt logtostderr=true blog.proto

