package rid_test

import (
	"fmt"

	"{{.D.ModuleName}}/internal/pkg/rid"
)

func ExampleResourceID_String() {
	// Define a resource identifier, for example, a user resource.
	userID := rid.UserID

	// Call the String method to convert the ResourceID type to a string type.
	idString := userID.String()

	// Output the result.
	fmt.Println(idString)

	// Output:
	// user
}
