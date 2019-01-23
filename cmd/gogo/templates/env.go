package templates

var (
	envTemplate = `#!/usr/bin/env bash

echo "***DEPRECATED***"
echo "Please use go mod instead!"
exit(1)

export GOGOROOT=$(pwd)

# adjust GOPATH
case ":$GOPATH:" in
    *":$GOGOROOT:"*) :;;
    *) GOPATH=$GOGOROOT:$GOPATH;;
esac
export GOPATH


# adjust PATH
readopts="ra"
if [ -n "$ZSH_VERSION" ]; then
    readopts="rA";
fi
while IFS=':' read -$readopts ARR; do
    for i in "${ARR[@]}"; do
        case ":$PATH:" in
            *":$i/bin:"*) :;;
            *) PATH=$i/bin:$PATH
        esac
    done
done <<< "$GOPATH"
export PATH


# mock development && test envs
if [ ! -d "$GOGOROOT/src/{{.Namespace}}/{{.Application}}" ];
then
    mkdir -p "$GOGOROOT/src/{{.Namespace}}"
    ln -s "$GOGOROOT/gogo/" "$GOGOROOT/src/{{.Namespace}}/{{.Application}}"
fi
`
)
