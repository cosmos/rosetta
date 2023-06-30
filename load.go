// nolint:SA1019 -- Since we are usinbg grpc_reflection_v1alpha which is deprecated, we excuse this rule
package rosetta

import (
	"context"
	reflectionv1beta1 "cosmossdk.io/api/cosmos/base/reflection/v1beta1"
	"crypto/tls"
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"io"
	"strings"
)

func ReflectInterfaces(ir codectypes.InterfaceRegistry, endpoint string) (err error) {
	ctx := context.Background()
	client, err := openClient(endpoint)
	if err != nil {
		return err
	}

	fdSet, err := getFileDescriptorSet(ctx, client)
	if err != nil {
		return err
	}

	fmt.Println("got fdset")
	for _, descriptorProto := range fdSet.File {
		if descriptorProto != nil {
			registerProtoInterface(ir, descriptorProto)
		}
	}
	return err
}

func openClient(endpoint string) (client *grpc.ClientConn, err error) {
	creds := credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS12,
	})

	client, err = grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		fmt.Println("[ERROR] getting grpc client connection")
		return nil, err
	}
	return client, err
}

func getFileDescriptorSet(c context.Context, client *grpc.ClientConn) (fdSet *descriptorpb.FileDescriptorSet, err error) {
	fdSet = &descriptorpb.FileDescriptorSet{}

	interfaceImplNames, err := getInterfaceImplNames(c, client)
	if err != nil {
		return fdSet, err
	}

	reflectClient, err := grpc_reflection_v1alpha.NewServerReflectionClient(client).ServerReflectionInfo(c)
	if err != nil {
		return fdSet, err
	}

	fdMap := map[string]*descriptorpb.FileDescriptorProto{}
	waitListServiceRes := make(chan *grpc_reflection_v1alpha.ListServiceResponse)
	wait := make(chan struct{})
	go func() {
		for {
			in, err := reflectClient.Recv()
			if err == io.EOF {
				close(wait)
				return
			}
			if err != nil {
				fmt.Println("[ERROR] Reflection failed on reflectClient:", err.Error())
				return
			}

			switch res := in.MessageResponse.(type) {
			case *grpc_reflection_v1alpha.ServerReflectionResponse_ErrorResponse:
				fmt.Println("[ERROR] Server reflection response:", res.ErrorResponse.String())
			case *grpc_reflection_v1alpha.ServerReflectionResponse_ListServicesResponse:
				waitListServiceRes <- res.ListServicesResponse
			case *grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse:
				for _, bz := range res.FileDescriptorResponse.FileDescriptorProto {
					fd := &descriptorpb.FileDescriptorProto{}
					err := proto.Unmarshal(bz, fd)
					if err != nil {
						fmt.Println("[ERROR] error happening while unmarshalling proto message", err.Error())
					}
					fdMap[fd.GetName()] = fd
				}
			}
		}
	}()

	if err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{},
	}); err != nil {
		fmt.Println("[ERROR] on ServerRefleciion services", err.Error())
	}

	listServiceRes := <-waitListServiceRes

	for _, response := range listServiceRes.Service {
		err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{
			MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: response.Name,
			},
		})
		if err != nil {
			fmt.Println("[ERROR] on ServerRefleciion services", err.Error())
		}
	}

	for _, msgName := range interfaceImplNames {
		err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{
			MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: msgName,
			},
		})
		if err != nil {
			fmt.Println("[ERROR] on getting interfaceImplNames", err.Error())
		}
	}

	if err = reflectClient.CloseSend(); err != nil {
		fmt.Println("[ERROR] on closing reflectClient", err.Error())
	}

	<-wait

	for _, descriptorProto := range fdMap {
		fdSet.File = append(fdSet.File, descriptorProto)
	}

	return fdSet, err
}

func getInterfaceImplNames(c context.Context, client *grpc.ClientConn) (interfaceImplNames []string, err error) {
	cosmosReflectBetaClient := reflectionv1beta1.NewReflectionServiceClient(client)
	interfacesRes, err := cosmosReflectBetaClient.ListAllInterfaces(c, &reflectionv1beta1.ListAllInterfacesRequest{})

	if err == nil {
		for _, iface := range interfacesRes.InterfaceNames {
			implRes, err := cosmosReflectBetaClient.ListImplementations(c, &reflectionv1beta1.ListImplementationsRequest{
				InterfaceName: iface,
			})
			if err == nil {
				interfaceImplNames = append(interfaceImplNames, cleanImplMsgNames(implRes.GetImplementationMessageNames())...)
			}
		}
	}
	return interfaceImplNames, err
}

func cleanImplMsgNames(implMessages []string) (cleanImplMessages []string) {
	for _, implMessage := range implMessages {
		cleanImplMessages = append(cleanImplMessages, implMessage[1:])
	}

	return cleanImplMessages
}

func registerProtoInterface(registry codectypes.InterfaceRegistry, fileDescriptor *descriptorpb.FileDescriptorProto) {
	name := "/" + strings.ReplaceAll(fileDescriptor.GetName(), "/", ".")
	descriptorMessageInterface := fileDescriptor.ProtoReflect().Interface()
	registry.RegisterInterface(name, &descriptorMessageInterface)
}
