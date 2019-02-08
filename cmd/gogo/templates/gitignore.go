package templates

var (
	gitIgnoreTemplate = `# Compiled Object files, Static and Dynamic libs (Shared Objects)
*.o
*.a
*.so
*.out

# Folders
_obj
_test
bin
pkg
src

# Architecture specific extensions/prefixes
*.[568vq]
[568vq].out

*.cgo1.go
*.cgo2.c
_cgo_defun.c
_cgo_gotypes.go
_cgo_export.*
_testmain.go

*.exe
*.test
*.prof

# development & test config files
*.development.yml
*.development.json
*.test.yml
*.test.json
`
)
