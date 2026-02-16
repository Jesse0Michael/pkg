package interceptors

import (
	"strings"

	options "github.com/jesse0michael/pkg/grpc/proto/options/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ResolveMethod parses a gRPC full method name and returns the service and
// method descriptors from the global proto registry.
func ResolveMethod(fullMethod string) (protoreflect.ServiceDescriptor, protoreflect.MethodDescriptor) {
	parts := strings.Split(strings.TrimPrefix(fullMethod, "/"), "/")
	if len(parts) != 2 {
		return nil, nil
	}

	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(parts[0]))
	if err != nil {
		return nil, nil
	}

	serviceDesc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, nil
	}

	return serviceDesc, serviceDesc.Methods().ByName(protoreflect.Name(parts[1]))
}

// MethodBoolOption reads a bool extension from the method's options.
func MethodBoolOption(md protoreflect.MethodDescriptor, ext *protoimpl.ExtensionInfo) bool {
	if md == nil {
		return false
	}
	opts, ok := md.Options().(*descriptorpb.MethodOptions)
	if !ok || opts == nil {
		return false
	}
	val, ok := proto.GetExtension(opts, ext).(bool)
	return ok && val
}

// ServiceBoolOption reads a bool extension from the service's options.
func ServiceBoolOption(sd protoreflect.ServiceDescriptor, ext *protoimpl.ExtensionInfo) bool {
	if sd == nil {
		return false
	}
	opts, ok := sd.Options().(*descriptorpb.ServiceOptions)
	if !ok || opts == nil {
		return false
	}
	val, ok := proto.GetExtension(opts, ext).(bool)
	return ok && val
}

// HasNoAuth returns true if the method or its parent service opts out of authentication.
func HasNoAuth(fullMethod string) bool {
	sd, md := ResolveMethod(fullMethod)
	return MethodBoolOption(md, options.E_NoAuth) || ServiceBoolOption(sd, options.E_ServiceNoAuth)
}

// HasAdminOnly returns true if the method or its parent service requires admin access.
func HasAdminOnly(fullMethod string) bool {
	sd, md := ResolveMethod(fullMethod)
	return MethodBoolOption(md, options.E_AdminOnly) || ServiceBoolOption(sd, options.E_ServiceAdminOnly)
}

// HasRejectReadOnly returns true if the method rejects read-only users.
func HasRejectReadOnly(fullMethod string) bool {
	_, md := ResolveMethod(fullMethod)
	return MethodBoolOption(md, options.E_RejectReadOnly)
}
