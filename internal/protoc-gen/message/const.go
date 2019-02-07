package message

const (
	// tag numbers in FileDescriptorProto
	packagePath = 2 // package
	messagePath = 4 // message_type
	enumPath    = 5 // enum_type
	servicePath = 6 // service
	// tag numbers in DescriptorProto
	messageFieldPath   = 2 // field
	messageMessagePath = 3 // nested_type
	messageEnumPath    = 4 // enum_type
	messageOneofPath   = 8 // oneof_decl
	// tag numbers in ServiceDescriptorProto
	serviceNamePath    = 1 // name
	serviceMethodPath  = 2 // method
	serviceOptionsPath = 3 // options
	// tag numbers in MethodDescriptorProto
	methodNamePath   = 1 // name
	methodInputPath  = 2 // input_type
	methodOutputPath = 3 // output_type
)
