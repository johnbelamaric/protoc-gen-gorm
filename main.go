package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

func main() {
	gen := generator.New()
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		gen.Error(err, "reading input")
	}
	if err = proto.Unmarshal(data, gen.Request); err != nil {
		gen.Error(err, "unmarshalling proto")
	}
	//fmt.Fprintf(os.Stderr,"%s", gen.Request.String())
	gen.CommandLineParameters(gen.Request.GetParameter())
	gen.WrapTypes()
	gen.SetPackageNames()
	gen.BuildTypeNameMap()
	plug := &ormPlugin{}
	gen.GeneratePlugin(plug)

	for i := 0; i < len(gen.Response.File); i++ {
		// Rename file type
		gen.Response.File[i].Name = proto.String(strings.Replace(*gen.Response.File[i].Name, ".pb.go", ".pb.orm.go", -1))
		// Put into subfolder
		if plug.newPackage != "" {
			gen.Response.File[i].Name = proto.String(fmt.Sprintf("%s/%s", plug.newPackage, *gen.Response.File[i].Name))

			content := *gen.Response.File[i].Content
			// Swap out the package name and package name in comment
			content = *proto.String(strings.Replace(content,
				fmt.Sprintf("package %s", plug.originalPackage),
				fmt.Sprintf("package %s", plug.newPackage), 1))
			content = *proto.String(strings.Replace(content,
				fmt.Sprintf("Package %s", plug.originalPackage),
				fmt.Sprintf("Package %s", plug.newPackage), 1))
			// For some reason, it autoimports the new package name
			content = *proto.String(strings.Replace(content, fmt.Sprintf(`import _ "%s"`, plug.newPackage), "", 1))
			gen.Response.File[i].Content = &content
		}
	}
	data, err = proto.Marshal(gen.Response)
	if err != nil {
		gen.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		gen.Error(err, "failed to write output proto")
	}
}
