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

import "log"

func panicIfError(e error) {
	if e != nil {
		panic(e)
	}
}

func indexOf(element string, data []string) int {
	for i, v := range data {
		if element == v {
			return i
		}
	}
	return -1
}

func logDebug(fmt string, v ...any) {
	if *logLevel > 0 {
		log.Printf(fmt, v...)
	}
}
