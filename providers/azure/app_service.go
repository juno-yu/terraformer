// Copyright 2020 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/Azure/go-autorest/autorest"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type AppServiceGenerator struct {
	AzureService
}

/*
func (g *AppServiceGenerator) listSQLDatabasesAndContainersBehind(resourceGroupName string, accountName string) ([]terraformutils.Resource, []terraformutils.Resource, error) {
	var resourcesDatabase []terraformutils.Resource
	var resourcesContainer []terraformutils.Resource
	ctx := context.Background()
	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	SQLResourcesClient := documentdb.NewSQLResourcesClient(subscriptionID, subscriptionID)
	SQLResourcesClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)

	sqlDatabases, err := SQLResourcesClient.ListSQLDatabases(ctx, resourceGroupName, accountName)
	if err != nil {
		return nil, nil, err
	}
	for _, sqlDatabase := range *sqlDatabases.Value {
		// NOTE:
		// For a similar reason as
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7472#issuecomment-650684349
		// The cosmosdb resource format change is NOT yet addressed in terraform provider
		// This line is a workaround to convert to old format, and might be removed if they deprecate the old format
		sqlDatabaseIDInOldFormat := strings.Replace(*sqlDatabase.ID, "sqlDatabases", "databases", 1)
		resourcesDatabase = append(resourcesDatabase, terraformutils.NewSimpleResource(
			sqlDatabaseIDInOldFormat,
			*sqlDatabase.Name,
			"azurerm_cosmosdb_sql_database",
			g.ProviderName,
			[]string{}))

		sqlContainers, err := SQLResourcesClient.ListSQLContainers(ctx, resourceGroupName, accountName, *sqlDatabase.Name)
		if err != nil {
			return nil, nil, err
		}
		for _, sqlContainer := range *sqlContainers.Value {
			// NOTE:
			// For a similar reason as
			// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7472#issuecomment-650684349
			// The cosmosdb resource format change is NOT yet addressed in terraform provider
			// This line is a workaround to convert to old format, and might be removed if they deprecate the old format
			sqlContainerIDInOldFormat := strings.Replace(*sqlContainer.ID, "sqlDatabases", "databases", 1)
			resourcesContainer = append(resourcesContainer, terraformutils.NewSimpleResource(
				sqlContainerIDInOldFormat,
				*sqlContainer.Name,
				"azurerm_cosmosdb_sql_container",
				g.ProviderName,
				[]string{}))
		}
	}

	return resourcesDatabase, resourcesContainer, nil
}

func (g *AppServiceGenerator) listTables(resourceGroupName string, accountName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	// NOTE:
	// there will be a parameter simplification for interface if we update the package
	// https://github.com/Azure/azure-sdk-for-go/blob/v42.0.0/services/cosmos-db/mgmt/2020-03-01/documentdb/tableresources.go#L35
	// https://github.com/Azure/azure-sdk-for-go/blob/v44.0.0/services/cosmos-db/mgmt/2020-03-01/documentdb/tableresources.go#L35
	TableResourcesClient := documentdb.NewTableResourcesClient(subscriptionID, subscriptionID)
	TableResourcesClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)

	tables, err := TableResourcesClient.ListTables(ctx, resourceGroupName, accountName)
	if err != nil {
		return nil, err
	}
	for _, table := range *tables.Value {
		resources = append(resources, terraformutils.NewSimpleResource(
			*table.ID,
			*table.Name,
			"azurerm_cosmosdb_table",
			g.ProviderName,
			[]string{}))
	}

	return resources, nil
}

*/
func (g *AppServiceGenerator) listAndAddForAppServices() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	subscriptionID := g.Args["config"].(authentication.Config).SubscriptionID
	// NOTE:
	// there will be a parameter simplification for interface if we update the package
	// https://github.com/Azure/azure-sdk-for-go/blob/v42.0.0/services/cosmos-db/mgmt/2020-03-01/documentdb/databaseaccounts.go#L35
	// https://github.com/Azure/azure-sdk-for-go/blob/v44.0.0/services/cosmos-db/mgmt/2020-03-01/documentdb/databaseaccounts.go#L35
	AppsClient := web.NewAppsClient(subscriptionID)
	AppsClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)
	appsClientResultIterator, err := AppsClient.ListComplete(ctx)
	if err != nil {
		return nil, err
	}

	for appsClientResultIterator.NotDone() {
		app := appsClientResultIterator.Value()
		appName := *app.Name
		resources = append(resources, terraformutils.NewSimpleResource(
			*app.ID,
			appName,
			"azurerm_app_service",
			"azurerm",
			[]string{}))
		if err := appsClientResultIterator.Next(); err != nil {
			log.Println(err)
			return nil, err
		}
		id, err := ParseAzureResourceID(*app.ID)
		if err != nil {
			return nil, err
		}

		slotResultIterator, err := AppsClient.ListSlotsComplete(ctx, id.ResourceGroup, appName)
		for slotResultIterator.NotDone() {
			slot := slotResultIterator.Value()
			resources = append(resources, terraformutils.NewSimpleResource(
				*slot.ID,
				*slot.Name,
				"azurerm_app_service_slot",
				"azurerm",
				[]string{}))
			if err := slotResultIterator.Next(); err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *AppServiceGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForAppServices,
	}

	for _, f := range functions {
		resources, err := f()
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}
