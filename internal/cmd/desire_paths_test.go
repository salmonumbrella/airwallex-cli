package cmd

import (
	"reflect"
	"testing"
)

func TestListResourceMap_BillingContactAliases(t *testing.T) {
	resourceMap := listResourceMap()

	tests := []struct {
		resource string
		want     []string
	}{
		{resource: "customers", want: []string{"billing", "customers", "list"}},
		{resource: "contacts", want: []string{"billing", "customers", "list"}},
	}

	for _, tt := range tests {
		got, ok := resourceMap[tt.resource]
		if !ok {
			t.Fatalf("resource %q not found in listResourceMap", tt.resource)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("listResourceMap[%q] = %v, want %v", tt.resource, got, tt.want)
		}
	}
}

func TestCreateResourceMap_BillingContactAliases(t *testing.T) {
	resourceMap := createResourceMap()

	tests := []struct {
		resource string
		want     []string
	}{
		{resource: "customer", want: []string{"billing", "customers", "create"}},
		{resource: "contact", want: []string{"billing", "customers", "create"}},
		{resource: "contacts", want: []string{"billing", "customers", "create"}},
	}

	for _, tt := range tests {
		got, ok := resourceMap[tt.resource]
		if !ok {
			t.Fatalf("resource %q not found in createResourceMap", tt.resource)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("createResourceMap[%q] = %v, want %v", tt.resource, got, tt.want)
		}
	}
}
