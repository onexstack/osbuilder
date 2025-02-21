// This file defines the Protobuf messages for managing {{.PluralName}}.
//
syntax = "proto3"; // Specifies the syntax version used in this file.

package {{.APIVersion}};

import "google/protobuf/timestamp.proto"; // Importing Google's timestamp type for date/time fields.

// Specifies the Go package for generated code.
option go_package = "{{.ModuleName}}/pkg/api/{{.ComponentName}}/{{.APIVersion}};{{.APIVersion}}";

// {{.SingularName}} represents a {{.SingularLower}} with its metadata.
message {{.SingularName}} {
    // {{.SingularName}}ID is the unique identifier for the {{.SingularLower}}.
    string {{.SingularLowerFirst}}ID = 1;
    // CreatedAt is the timestamp when the {{.SingularLower}} was created.
    google.protobuf.Timestamp createdAt = 2;
    // UpdatedAt is the timestamp when the {{.SingularLower}} was last updated.
    google.protobuf.Timestamp updatedAt = 3;
    // TODO: Add additional fields if needed.
}

// Create{{.SingularName}}Request represents the request message for creating a new {{.SingularLower}}.
message Create{{.SingularName}}Request {
    // TODO: Add additional fields if needed.
}

// Create{{.SingularName}}Response represents the response message for a successful {{.SingularLower}} creation.
message Create{{.SingularName}}Response {
    // {{.SingularName}}ID is the unique identifier of the newly created {{.SingularLower}}.
    string {{.SingularLowerFirst}}ID = 1;
    // TODO: Add additional fields to return if needed.
}

// Update{{.SingularName}}Request represents the request message for updating an existing {{.SingularLower}}.
message Update{{.SingularName}}Request {
    // {{.SingularName}}ID is the unique identifier of the {{.SingularLower}} to update.
    string {{.SingularLowerFirst}}ID = 1;
    // TODO: Add additional fields to update if needed.
}

// Update{{.SingularName}}Response represents the response message for a successful {{.SingularLower}} update.
message Update{{.SingularName}}Response {
    // TODO: Add additional fields to return if needed.
}

// Delete{{.SingularName}}Request represents the request message for deleting one or more {{.PluralLower}}.
message Delete{{.SingularName}}Request {
    // {{.SingularName}}IDs is the list of unique identifiers for the {{.PluralLower}} to delete.
    repeated string {{.SingularLowerFirst}}IDs = 1;
    // TODO: Add additional fields if needed.
}

// Delete{{.SingularName}}Response represents the response message for a successful {{.SingularLower}} deletion.
message Delete{{.SingularName}}Response {
    // TODO: Add additional fields to return if needed.
}

// Get{{.SingularName}}Request represents the request message for retrieving a specific {{.SingularLower}}.
message Get{{.SingularName}}Request {
    // {{.SingularName}}ID is the unique identifier of the {{.SingularLower}} to retrieve.
    // @gotags: uri:"{{.SingularLowerFirst}}ID"
    string {{.SingularLowerFirst}}ID = 1;
}

// Get{{.SingularName}}Response represents the response message for a successful retrieval of a {{.SingularLower}}.
message Get{{.SingularName}}Response {
    // {{.SingularName}} is the retrieved {{.SingularLower}} object.
    {{.SingularName}} {{.SingularLowerFirst}} = 1;
}

// List{{.SingularName}}Request represents the request message for listing {{.PluralLower}}
// with pagination and optional filters.
message List{{.SingularName}}Request {
    // Offset is the starting point of the list for pagination.
    // @gotags: form:"offset"
    int64 offset = 1;
    // Limit is the maximum number of {{.PluralLower}} to return.
    // @gotags: form:"limit"
    int64 limit = 2;
    // TODO: Add additional query fields if needed.
}

// List{{.SingularName}}Response represents the response message for listing {{.PluralLower}}.
message List{{.SingularName}}Response {
    // TotalCount is the total number of {{.PluralLower}} matching the query.
    int64 total = 1;
    // {{.SingularName}} is the list of {{.PluralLower}} in the current page.
    repeated {{.SingularName}} {{.PluralLowerFirst}} = 2;
}
