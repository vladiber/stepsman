/*
Copyright © 2020 stepsman authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/fortify500/stepsman/bl"
	"github.com/spf13/cobra"
)

var fileName string

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a run and move the cursor to the first step",
	Long: `Start a run and move the cursor to the first step.
You can specify either a file or stdin (stdin not implemented yet)`,
	Run: func(cmd *cobra.Command, args []string) {
		var r bl.Run
		fmt.Println("start called")
		if len(fileName) == 0 {
			fmt.Println("you must specify a file name")
			return
		}
		err := r.Start(fileName)
		if err!=nil {
			fmt.Println("failed to start: %w", err)
			return
		}
		fmt.Println("run: %#v", r)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringVarP(&fileName, "file", "f", "", "File to start run")
	startCmd.MarkFlagRequired("file")
}
