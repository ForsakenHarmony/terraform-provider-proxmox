// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package registry

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// We use a global registry to be able to test resources and dataSources in the same directory / package
// as they are defined.
type registry struct {
	registrationClosed bool

	dataSources []func() datasource.DataSource
	resources   []func() resource.Resource

	mutex sync.Mutex
}

//nolint:gochecknoglobals
var reg = registry{}

// // AddDataSourceFactory registers the specified data source type name and factory.
// func AddDataSourceFactory(factory func() datasource.DataSource) {
// 	reg.mutex.Lock()
// 	defer reg.mutex.Unlock()
//
// 	if reg.registrationClosed {
// 		panic("Data Source registration is closed")
// 	}
//
// 	reg.dataSources = append(reg.dataSources, factory)
// }

// AddResourceFactory registers the specified resource type name and factory.
func AddResourceFactory(factory func() resource.Resource) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.registrationClosed {
		panic("Resource registration is closed")
	}

	reg.resources = append(reg.resources, factory)
}

// DataSourceFactories returns the registered data source factories.
// Data Source registration is closed.
func DataSourceFactories() []func() datasource.DataSource {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	reg.registrationClosed = true

	return reg.dataSources
}

// ResourceFactories returns the registered resource factories.
// Resource registration is closed.
func ResourceFactories() []func() resource.Resource {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	reg.registrationClosed = true

	return reg.resources
}
