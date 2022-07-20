package gencli

import (
	"fmt"
	"path"
	"sort"
	"strconv"

	"go.einride.tech/protoc-gen-go-cli/cli"
	"google.golang.org/protobuf/compiler/protogen"
)

func GenerateRootCommandFile(gen *protogen.Plugin, config cli.CompilerConfig) error {
	module, ok := getModuleParam(gen)
	if !ok {
		return fmt.Errorf("param root requires param module to be provided")
	}
	g := gen.NewGeneratedFile(path.Join(module, "root.go"), "")
	generateGeneratedFileHeader(g, gen)
	g.P("package main")
	cobraCommand := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "github.com/spf13/cobra",
		GoName:       "Command",
	})
	cliConfig := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/protoc-gen-go-cli/cli",
		GoName:       "Config",
	})
	g.P()
	g.P("func NewRootCommand() *", cobraCommand, " {")
	g.P("cmd := &", cobraCommand, "{")
	g.P("Use: ", strconv.Quote(config.Root), ",")
	g.P("}")
	servicesByName := getServicesByName(gen)
	type fileService struct {
		f *protogen.File
		s *protogen.Service
	}
	services := []*fileService{}
	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}
		for _, service := range file.Services {
			services = append(services, &fileService{
				f: file,
				s: service,
			})
		}
	}
	sort.Slice(services, func(i int, j int) bool {
		return services[i].s.GoName < services[j].s.GoName
	})
	for _, service := range services {
		newCommandFunction := g.QualifiedGoIdent(protogen.GoIdent{
			GoImportPath: service.f.GoImportPath,
			GoName:       "New" + service.s.GoName + "Command",
		})
		serviceCommand := getServiceCommandUse(servicesByName, service.s)
		g.P("cmd.AddCommand(", newCommandFunction, "(", strconv.Quote(serviceCommand), "))")
	}
	g.P("return cmd")
	g.P("}")
	g.P()
	g.P("func NewConfig() *", cliConfig, " {")
	g.P("return &", cliConfig, "{")
	g.P("Compiler: ", fmt.Sprintf("%#v", config), ",")
	g.P("}")
	g.P("}")
	return nil
}
