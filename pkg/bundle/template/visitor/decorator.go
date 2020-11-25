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

package visitor

import (
	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// -----------------------------------------------------------------------------

// InfrastructureDecorator decorates an bundle template infrastructure namespace
// to make it visitable.
func InfrastructureDecorator(entity *bundlev1.InfrastructureNS) InfrastructureAcceptor {
	return &infrastructureDecorator{
		entity: entity,
	}
}

type infrastructureDecorator struct {
	entity *bundlev1.InfrastructureNS
}

func (w *infrastructureDecorator) Accept(v InfrastructureVisitor) {
	v.VisitForProvider(w.entity)
}

// -----------------------------------------------------------------------------

// PlatformDecorator decorates an bundle template platform namespace
// to make it visitable.
func PlatformDecorator(entity *bundlev1.PlatformRegionNS) PlatformAcceptor {
	return &platformDecorator{
		entity: entity,
	}
}

type platformDecorator struct {
	entity *bundlev1.PlatformRegionNS
}

func (w *platformDecorator) Accept(v PlatformVisitor) {
	v.VisitForRegion(w.entity)
}

// -----------------------------------------------------------------------------

// ProductDecorator decorates an bundle template product namespace
// to make it visitable.
func ProductDecorator(entity *bundlev1.ProductComponentNS) ProductAcceptor {
	return &productDecorator{
		entity: entity,
	}
}

type productDecorator struct {
	entity *bundlev1.ProductComponentNS
}

func (w *productDecorator) Accept(v ProductVisitor) {
	v.VisitForComponent(w.entity)
}

// -----------------------------------------------------------------------------

// ApplicationDecorator decorates an bundle template application namespace
// to make it visitable.
func ApplicationDecorator(entity *bundlev1.ApplicationComponentNS) ApplicationAcceptor {
	return &applicationDecorator{
		entity: entity,
	}
}

type applicationDecorator struct {
	entity *bundlev1.ApplicationComponentNS
}

func (w *applicationDecorator) Accept(v ApplicationVisitor) {
	v.VisitForComponent(w.entity)
}
