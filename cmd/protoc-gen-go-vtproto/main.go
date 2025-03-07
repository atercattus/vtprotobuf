package main

import (
	"flag"
	"fmt"
	"strings"

	_ "github.com/atercattus/vtprotobuf/features/equal"
	_ "github.com/atercattus/vtprotobuf/features/grpc"
	_ "github.com/atercattus/vtprotobuf/features/marshal"
	_ "github.com/atercattus/vtprotobuf/features/pool"
	_ "github.com/atercattus/vtprotobuf/features/size"
	_ "github.com/atercattus/vtprotobuf/features/unmarshal"
	"github.com/atercattus/vtprotobuf/generator"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type ObjectSet map[protogen.GoIdent]bool

func (o ObjectSet) String() string {
	return fmt.Sprintf("%#v", o)
}

func (o ObjectSet) Set(s string) error {
	idx := strings.LastIndexByte(s, '.')
	if idx < 0 {
		return fmt.Errorf("invalid object name: %q", s)
	}

	ident := protogen.GoIdent{
		GoImportPath: protogen.GoImportPath(s[0:idx]),
		GoName:       s[idx+1:],
	}
	o[ident] = true
	return nil
}

func main() {
	var features string
	poolable := make(ObjectSet)

	var f flag.FlagSet
	f.Var(poolable, "pool", "use memory pooling for this object")
	f.StringVar(&features, "features", "all", "list of features to generate (separated by '+')")

	protogen.Options{ParamFunc: f.Set}.Run(func(plugin *protogen.Plugin) error {
		return generateAllFiles(plugin, strings.Split(features, "+"), poolable)
	})
}

var SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

func generateAllFiles(plugin *protogen.Plugin, featureNames []string, poolable ObjectSet) error {
	ext := &generator.Extensions{Poolable: poolable}
	gen, err := generator.NewGenerator(plugin.Files, featureNames, ext)
	if err != nil {
		return err
	}

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		gf := plugin.NewGeneratedFile(file.GeneratedFilenamePrefix+"_vtproto.pb.go", file.GoImportPath)
		if !gen.GenerateFile(gf, file) {
			gf.Skip()
		}
	}

	plugin.SupportedFeatures = SupportedFeatures
	return nil
}
