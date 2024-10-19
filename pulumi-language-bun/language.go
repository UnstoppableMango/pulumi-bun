package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/unstoppablemango/pulumi-bun/pulumi-language-bun/proto/pulumi"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LanguageServer struct{}

// About implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) About(context.Context, *connect.Request[pulumi.AboutRequest]) (*connect.Response[pulumi.AboutResponse], error) {
	panic("unimplemented")
}

// GeneratePackage implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) GeneratePackage(context.Context, *connect.Request[pulumi.GeneratePackageRequest]) (*connect.Response[pulumi.GeneratePackageResponse], error) {
	panic("unimplemented")
}

// GenerateProgram implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) GenerateProgram(context.Context, *connect.Request[pulumi.GenerateProgramRequest]) (*connect.Response[pulumi.GenerateProgramResponse], error) {
	panic("unimplemented")
}

// GenerateProject implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) GenerateProject(context.Context, *connect.Request[pulumi.GenerateProjectRequest]) (*connect.Response[pulumi.GenerateProjectResponse], error) {
	panic("unimplemented")
}

// GetPluginInfo implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) GetPluginInfo(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[pulumi.PluginInfo], error) {
	panic("unimplemented")
}

// GetProgramDependencies implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) GetProgramDependencies(context.Context, *connect.Request[pulumi.GetProgramDependenciesRequest]) (*connect.Response[pulumi.GetProgramDependenciesResponse], error) {
	panic("unimplemented")
}

// GetRequiredPlugins implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) GetRequiredPlugins(context.Context, *connect.Request[pulumi.GetRequiredPluginsRequest]) (*connect.Response[pulumi.GetRequiredPluginsResponse], error) {
	panic("unimplemented")
}

// InstallDependencies implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) InstallDependencies(context.Context, *connect.Request[pulumi.InstallDependenciesRequest], *connect.ServerStream[pulumi.InstallDependenciesResponse]) error {
	panic("unimplemented")
}

// Pack implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) Pack(context.Context, *connect.Request[pulumi.PackRequest]) (*connect.Response[pulumi.PackResponse], error) {
	panic("unimplemented")
}

// Run implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) Run(context.Context, *connect.Request[pulumi.RunRequest]) (*connect.Response[pulumi.RunResponse], error) {
	panic("unimplemented")
}

// RunPlugin implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) RunPlugin(context.Context, *connect.Request[pulumi.RunPluginRequest], *connect.ServerStream[pulumi.RunPluginResponse]) error {
	panic("unimplemented")
}

// RuntimeOptionsPrompts implements pulumiconnect.LanguageRuntimeHandler.
func (l *LanguageServer) RuntimeOptionsPrompts(context.Context, *connect.Request[pulumi.RuntimeOptionsRequest]) (*connect.Response[pulumi.RuntimeOptionsResponse], error) {
	panic("unimplemented")
}
