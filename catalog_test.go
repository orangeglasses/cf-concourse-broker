package main

import "testing"

func TestCatalogLoad(t *testing.T) {
	jsonFilePath := "./catalog.json"
	services, err := CatalogLoad(jsonFilePath)
	if err != nil {
		t.Error("Unable to load catalog from json file:" + jsonFilePath)
	} else {
		if len(services) == 0 {
			t.Error("Service Array is empy.")
		}
		if services[0].ID == "" {
			t.Error("No ID found for first service")
		}
	}
}

func TestCatalogLoadFail(t *testing.T) {
	jsonFilePath := "./nonexistentfile"
	services, err := CatalogLoad(jsonFilePath)
	if err == nil {
		t.Error("CatalogLoad didn't raise an error when loading a non-existant json file")
	}
	if len(services) > 0 {
		t.Error("Catalog array was populated even when a nonexistent json was loaded")
	}
}
