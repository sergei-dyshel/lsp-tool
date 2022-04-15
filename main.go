// Copyright 2018 Jacob Dufault
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

var progName = path.Base(os.Args[0])

var help = fmt.Sprintf(`LSP server wrapper

Filtering (--enable/--disable) allows to run multiple servers so that
with only one of them providing specific capabilities.

PROVIDERS are language server capabilities without the "Provider" at the end
	(see https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#serverCapabilities for complete list)

Examples:
  %[1]v --enable completion,codeAction clangd
  %[1]v --disable completion,codeAction ccls
`, progName)

var (
	enableProviders  *[]string
	disableProviders *[]string
	logFileName      *string
	logLevel         *int
)

func validateArgs(_ *cobra.Command, args []string) error {
	if len(*enableProviders) > 0 && len(*disableProviders) > 0 {
		return errors.New("both enable/disable flags given")
	}
	if len(args) == 0 {
		return errors.New("empty command")
	}
	return nil
}

func run(_ *cobra.Command, args []string) error {
	if *logFileName == "" {
		log.SetOutput(os.Stderr)
		log.SetPrefix("lsp-tool: ")
	} else {
		file, err := os.OpenFile(*logFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("can not open log file: %w", err)
		}
		log.SetOutput(file)
	}
	log.SetFlags(0)

	log.Printf("Running %s", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	lsStdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not create stdout pipe: %w", err)
	}
	err = cmd.Start()
	panicIfError(err) // should not happen

	var mode filterMode = noFilter
	var providers []string
	if len(*enableProviders) > 0 {
		mode = enableFilter
		providers = *enableProviders
	} else if len(*disableProviders) > 0 {
		mode = disableFilter
		providers = *disableProviders
	}

	go stdoutReader(lsStdout, mode, providers)

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

var (
	rootCmd = &cobra.Command{
		Use:  fmt.Sprintf("%s [-l] [-v] [ -e ... | -d ... ] -- <exe> <args>...", progName),
		Long: help,
		RunE: run,
		Args: validateArgs,

		DisableFlagsInUseLine: true,
	}
)

func main() {
	enableProviders = rootCmd.Flags().StringSliceP("enable", "e", nil, "allow only the providers from `PROVIDERS` (comma-separated)")
	disableProviders = rootCmd.Flags().StringSliceP("disable", "d", nil, "allow all providers except from `PROVIDERS` (comma-separated)")
	logFileName = rootCmd.Flags().StringP("log", "l", "", "write log to `FILENAME` (default: stderr)")
	logLevel = rootCmd.Flags().CountP("verbose", "v", "Verbosity level, repeat to increase verbosity")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
