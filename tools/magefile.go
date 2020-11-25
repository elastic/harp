// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// +build mage

package main

import (
	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

func Build() {
	color.Red("# Installing tools ---------------------------------------------------------")
	mg.SerialDeps(Go.Vendor, Go.Tools)
}

type Go mg.Namespace

var deps = []string{
	"github.com/izumin5210/gex/cmd/gex",
}

// Vendor create tools vendors
func (Go) Vendor() error {
	color.Blue("## Vendoring dependencies")
	return sh.RunV("go", "mod", "vendor")
}

// Tools updates tools from package
func (Go) Tools() error {
	color.Blue("## Installing tools")
	return sh.RunV("go", "run", "github.com/izumin5210/gex/cmd/gex", "--build")
}
