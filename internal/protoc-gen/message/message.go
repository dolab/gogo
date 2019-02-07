package message

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// A MessageDefinition is parsed protoc file descriptor
type MessageDefinition struct {
	// Parent is the parent message, if this was defined as a nested message. If
	// this was defiend at the top level, parent is nil.
	Parent *MessageDefinition

	// Descriptor is is the DescriptorProto defining the message.
	Descriptor *descriptor.DescriptorProto

	// File is the File that the message was defined in. Or, if it has been
	// publicly imported, what File was that import performed in?
	File *descriptor.FileDescriptorProto

	// Comments describes the comments surrounding a message's definition. If it
	// was publicly imported, then these comments are from the actual source file,
	// not the file that the import was performed in.
	Comments DefinitionComments

	// path is the 'SourceCodeInfo' path. See the documentation for
	// github.com/golang/protobuf/protoc-gen-go/descriptor.SourceCodeInfo for an
	// explanation of its format.
	path []int32
}

// NewMessageFromFile gathers a mapping of fully-qualified protobuf names to
// their definitions. It scans a singles file at a time. It requires a mapping
// of .proto file names to their definitions in order to correctly handle
// 'import public' declarations; this mapping should include all files
// transitively imported by file.
func NewMessageFromFile(
	file *descriptor.FileDescriptorProto,
	nameToFiles map[string]*descriptor.FileDescriptorProto) map[string]*MessageDefinition {
	protoToMessages := make(map[string]*MessageDefinition)

	// First, gather all the messages defined at the top level.
	for i, descriptor := range file.MessageType {
		path := []int32{messagePath, int32(i)}
		message := &MessageDefinition{
			Parent:     nil,
			Descriptor: descriptor,
			File:       file,
			Comments:   commentsAtPath(path, file),
			path:       path,
		}

		protoToMessages[message.ProtoName()] = message

		// Next, all nested message definitions.
		for _, child := range message.descendants() {
			protoToMessages[child.ProtoName()] = child
		}
	}

	// Finally, all messages imported publicly.
	for _, depIdx := range file.PublicDependency {
		depFilename := file.Dependency[depIdx]
		depFile := nameToFiles[depFilename]
		depMessage := NewMessageFromFile(depFile, nameToFiles)
		for _, message := range depMessage {
			imported := &MessageDefinition{
				Parent:     message.Parent,
				Descriptor: message.Descriptor,
				File:       file,
				Comments:   commentsAtPath(message.path, depFile),
				path:       message.path,
			}

			protoToMessages[imported.ProtoName()] = imported
		}
	}

	return protoToMessages
}

// ProtoName returns the dot-delimited, fully-qualified protobuf name of the
// message.
func (m *MessageDefinition) ProtoName() string {
	prefix := "."
	if pkg := m.File.GetPackage(); pkg != "" {
		prefix += pkg + "."
	}

	if lineage := m.Lineage(); len(lineage) > 0 {
		for _, parent := range lineage {
			prefix += parent.Descriptor.GetName() + "."
		}
	}

	return prefix + m.Descriptor.GetName()
}

// Lineage returns m's parental chain all the way back up to a top-level message
// definition. The first element of the returned slice is the highest-level
// parent.
func (m *MessageDefinition) Lineage() []*MessageDefinition {
	var parents []*MessageDefinition
	for p := m.Parent; p != nil; p = p.Parent {
		parents = append([]*MessageDefinition{p}, parents...)
	}
	return parents
}

// descendants returns all the submessages defined within m, and all the
// descendants of those, recursively.
func (m *MessageDefinition) descendants() []*MessageDefinition {
	descendants := make([]*MessageDefinition, 0)
	for i, child := range m.Descriptor.NestedType {
		path := append(m.path, []int32{messageMessagePath, int32(i)}...)
		message := &MessageDefinition{
			Parent:     m,
			Descriptor: child,
			File:       m.File,
			Comments:   commentsAtPath(path, m.File),
			path:       path,
		}
		descendants = append(descendants, message)
		descendants = append(descendants, message.descendants()...)
	}

	return descendants
}

// DefinitionComments contains the comments surrounding a definition in a
// protobuf file.
//
// These follow the rules described by protobuf:
//
// A series of line comments appearing on consecutive lines, with no other
// tokens appearing on those lines, will be treated as a single comment.
//
// leading_detached_comments will keep paragraphs of comments that appear
// before (but not connected to) the current element. Each paragraph,
// separated by empty lines, will be one comment element in the repeated
// field.
//
// Only the comment content is provided; comment markers (e.g. //) are
// stripped out.  For block comments, leading whitespace and an asterisk
// will be stripped from the beginning of each line other than the first.
// Newlines are included in the output.
//
// Examples:
//
//   optional int32 foo = 1;  // Comment attached to foo.
//   // Comment attached to bar.
//   optional int32 bar = 2;
//
//   optional string baz = 3;
//   // Comment attached to baz.
//   // Another line attached to baz.
//
//   // Comment attached to qux.
//   //
//   // Another line attached to qux.
//   optional double qux = 4;
//
//   // Detached comment for corge. This is not leading or trailing comments
//   // to qux or corge because there are blank lines separating it from
//   // both.
//
//   // Detached comment for corge paragraph 2.
//
//   optional string corge = 5;
//   /* Block comment attached
//    * to corge.  Leading asterisks
//    * will be removed. */
//   /* Block comment attached to
//    * grault. */
//   optional int32 grault = 6;
//
//   // ignored detached comments.
type DefinitionComments struct {
	Leading         string
	Trailing        string
	LeadingDetached []string
}

func commentsAtPath(path []int32, sourceFile *descriptor.FileDescriptorProto) DefinitionComments {
	if sourceFile.SourceCodeInfo == nil {
		// The compiler didn't provide us with comments.
		return DefinitionComments{}
	}

	for _, loc := range sourceFile.SourceCodeInfo.Location {
		if pathEqual(path, loc.Path) {
			return DefinitionComments{
				Leading:         loc.GetLeadingComments(),
				LeadingDetached: loc.GetLeadingDetachedComments(),
				Trailing:        loc.GetTrailingComments(),
			}
		}
	}
	return DefinitionComments{}
}

func pathEqual(path1, path2 []int32) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i, v := range path1 {
		if path2[i] != v {
			return false
		}
	}
	return true
}
