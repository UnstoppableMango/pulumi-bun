package main

import (
	"fmt"
	"io/fs"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/spf13/afero"
)

// https://github.com/pulumi/pulumi/blob/fd07f42c6116ea9685b24ca35ba5832994196fc5/sdk/nodejs/cmd/pulumi-language-nodejs/main.go#L378

func getPlugins(fsys afero.Fs, root string) ([]*pulumirpc.PluginDependency, error) {
	plugins := []*pulumirpc.PluginDependency{}
	err := afero.Walk(fsys, root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking fs: %w", err)
	}

	return plugins, nil
}
