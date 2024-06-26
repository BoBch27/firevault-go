# Firevault
Firevault is a [Firestore](https://cloud.google.com/firestore/) object modeling tool to make life easier for Go devs. Inspired by [Firevault.js](https://github.com/bobch27/firevault-js).

Installation
------------
Use go get to install Firevault.

```go
go get github.com/bobch27/firevault-go
```

Importing
------------
Import the package in your code.

```go
import "github.com/bobch27/firevault-go"
```

Connection
------------
You can connect to Firevault using the [Firebase Admin SDK](https://firebase.google.com/docs/admin/setup) and the `Connect` method.

```go
import (
	"log"

	firebase "firebase.google.com/go"
	"github.com/bobch27/firevault-go"
)

ctx := context.Background()
app, err := firebase.NewApp(ctx, nil)
if err != nil {
  log.Fatalln("Firebase initialisation failed:", err)
}

client, err := app.Firestore(ctx)
if err != nil {
  log.Fatalln("Firestore initialisation failed:", err)
}

defer client.Close()

connection := firevault.Connect(client)
```

```go
import (
	"log"

	"cloud.google.com/go/firestore"
	"github.com/bobch27/firevault-go"
)

// Sets your Google Cloud Platform project ID.
projectId := "YOUR_PROJECT_ID"
ctx := context.Background()

client, err := firestore.NewClient(ctx, projectId)
if err != nil {
  log.Fatalln("Firestore initialisation failed:", err)
}

defer client.Close()

connection := firevault.Connect(client)
```

Models
------------
Defining a model is as simple as creating a struct with Firevault tags.

```go
type User struct {
	Name     string   `firevault:"name,required,omitempty"`
	Email    string   `firevault:"email,required,email,isUnique,omitempty"`
	Password string   `firevault:"password,required,min=6,transform=hashPass,omitempty"`
	Address  *Address `firevault:"address,omitempty"`
	Age      int      `firevault:"age,required,min=18,omitempty"`
}

type Address struct {
	Line1 string `firevault:",omitempty"`
	City  string `firevault:"-"`
}
```

Tags
------------
When defining a new struct type with Firevault tags, note that the tags' order matters (apart from the `omitempty` and `omitemptyupdate` tags, which can be used anywhere). 

The first tag is always the **field name** which will be used in Firestore. You can skip that by just using a comma, before adding further tags.

After that, each tag is a different validation rule, and they will be parsed in order.

Other than the validation tags, Firevault supports the following built-in tags:
- `omitempty` - If the field is set to it’s default value (e.g. `0` for `int`, or `""` for `string`), the field will be omitted from validation and Firestore.
- `omitemptyupdate` - Works the same way as `omitempty`, but only for the `Validate` and `UpdateById` methods. Ignored during `Create`.
- `-` - Ignores the field.

Validations
------------
Firevault validates fields' values based on the defined rules. There are **4** built-in validations, whilst also supporting the ability to add **custom** ones. 

*Again, the order in which they are executed depends on the tag order.*

*Built-in validations:*
- `required` - Validates whether the field's value is not the default type value (i.e. `nil` for `pointer`, `""` for `string`, `0` for `int` etc.). Fails when it is the default.
- `max` - Validates whether the field's value, or length, is less than or equal to the param's value. Requires a param (e.g. `max=20`). For numbers, it checks the value, for strings, maps and slices, it checks the length.
- `min` - Validates whether the field's value, or length, is greater than or equal to the param's value. Requires a param (e.g. `min=20`). For numbers, it checks the value, for strings, maps and slices, it checks the length.
- `email` - Validates whether the field's string value is a valid email address.

*Custom validations:*
- To define a custom validation, use the `connection`'s `RegisterValidation` method.
	- *Expects*:
		- name: A `string` defining the validation name
		- func: A function of type `ValidationFn`. The passed in function accepts two parameters. 
			- field: A `reflect.Value` of the field.
			- param: A `string` which will be validated against.
	- *Returns*:
		- result: A `bool` which returns `true` if check has passed, and `false` if it hasn't.

```go
connection.RegisterValidation("isUpper", func(fieldValue reflect.Value, _ string) bool {
	if fieldValue.Kind() != reflect.String {
		return false
	}

	s := fieldValue.String()
	return s == strings.toUpper(s)
})
```

You can then chain the tag like a normal one.

```go
type User struct {
	Name string `firevault:"name,required,isUpper,omitempty"`
}
```

Transformations
------------
Firevault also supports rules that transform the field's value. To use them, it's as simple as registering a transformation and adding a prefix to the tag.

- To define a transformation, use the `connection`'s `RegisterTransformation` method.
	- *Expects*:
		- name: A `string` defining the validation name
		- func: A function of type `TransformationFn`. The passed in function accepts one parameter. 
			- field: A `reflect.Value` of the field.
	- *Returns*:
		- newVal: An `interface{}` with the new value. 
		- error: An `error` in case something goes wrong during the transformation.

```go
connection.RegisterTransformation("toLower", func toUpper(fieldValue reflect.Value) (interface{}, error) {
	if fieldValue.Kind() != reflect.String {
		return fieldValue.Interface(), nil
	}

	if fieldValue.String() != "" {
		return strings.ToLower(fieldValue.String()), nil
	}

	return fieldValue.String(), nil
})
```

You can then chain the tag like a normal one, but don't forget to use the `transform=` prefix.

*Again, the tag order matters. Defining a transformation at the end, means the value will be updated **after** the validations, whereas a definition at the start, means the field will be updated and **then** validated.*

```go
type User struct {
	Email string `firevault:"email,required,email,transform=toLower,omitempty"`
}
```

Collections
------------
To create a collection instance, call the `NewCollection` method, using the struct type parameter, and passing in the `connection` instance, as well as a collection **name**.

```go
collection, err := firevault.NewCollection[User](connection, "users")
if err != nil {
	fmt.Println(err)
}
```

Methods
------------
The collection instance has **6** built-in methods to support interaction with Firestore.

- `Create` - A method which validates passed in data and adds it as a document to Firestore. 
	- *Expects*:
		- ctx: A context.
		- data: A `pointer` of a `struct` with populated fields which will be added to Firestore after validation.
		- options *(optional)*: An instance of `CreationOptions` with three properties. 
			- SkipRequired: A `bool` which when `true`, means the `required` tag will be ignored (i.e. the `required` check is skipped). Default value is `false`.
			- SkipValidation: A `bool` which when `true`, means all validation tags will be ingored (the `name` and `omitempty` tags will be acknowledged). Default is `false`.
			- ID: A `string` which will add a document to Firestore with the specified ID.
			- AllowEmptyFields: An optional `slice` of type `FieldPath`, which is used to specify which fields can ignore the `omitempty` tag. This can be useful when a field must be set to its zero value only on certain method calls. If left empty, all fields will honour the tag.
	- *Returns*:
		- id: A `string` with the new document's ID.
		- error: An `error` in case something goes wrong during validation or interaction with Firestore.
```go
user := &User{
	Name: 	  "Bobby Donev",
	Email:    "hello@bobbydonev.com",
	Password: "12356",
	Age:      26,
	Address:  &Address{
		Line1: "1 High Street",
		City:  "London",
	},
}
id, err := collection.Create(ctx, user)
if err != nil {
	fmt.Println(err)
} 
fmt.Println(id) // "6QVHL46WCE680ZG2Xn3X"
```
```go
id, err := collection.Create(ctx, user, CreationOptions{
	SkipRequired: true, 
	ID: "custom-id",
})
if err != nil {
	fmt.Println(err)
} 
fmt.Println(id) // "custom-id"
```
```go
user := User{
	Name: 	  "Bobby Donev",
	Email:    "hello@bobbydonev.com",
	Password: "12356",
	Age:      0,
	Address:  &Address{
		Line1: "1 High Street",
		City:  "London",
	},
}
id, err := collection.Create(ctx, &user, CreationOptions{
	AllowEmptyFields: []FieldPath{[]string{"age"}},
})
if err != nil {
	fmt.Println(err)
} 
fmt.Println(id) // "6QVHL46WCE680ZG2Xn3X"
```
- `UpdateById` - A method which validates passed in data and updates given Firestore document. 
	- *Expects*:
		- ctx: A context.
		- id: A `string` with the document's ID.
		- data: A `pointer` of a `struct` with populated fields which will be used to update the document after validation.
		- options *(optional)*: An instance of `UpdatingOptions` with three properties. 
			- SkipRequired: A `bool` which when `false`, means the `required` tag will not be ignored (i.e. the `required` check is not skipped). Default value is `true`.
			- SkipValidation: A `bool` which when `true`, means all validation tags will be ingored (the `name` and `omitempty` tags will be acknowledged). Default is `false`.
			- MergeFields: An optional `slice` of type `FieldPath`, which is used to specify which fields to be overwritten. Other fields on the document will be untouched. If left empty, all the fields given in the data argument will be overwritten.
			- AllowEmptyFields: An optional `slice` of type `FieldPath`, which is used to specify which fields can ignore the `omitempty` and `omitemptyupdate` tags. This can be useful when a field must be set to its zero value only on certain updates. If left empty, all fields will honour the two tags.
	- *Returns*:
		- error: An `error` in case something goes wrong during validation or interaction with Firestore.
	- ***Important***: 
		- If neither `omitempty`, nor `omitemptyupdate` tags have been used, non-specified field values in the passed in data will be set to Go's default values, thus updating all document fields. To prevent that behaviour, please use one of the two tags. 
		- If a document with the specified ID does not exist, Firestore will create one with the specified fields, so it's worth checking whether the doc exists before using the method.
```go
user := &User{
	Password: "123567",
}
err := collection.UpdateById(ctx, "6QVHL46WCE680ZG2Xn3X", user)
if err != nil {
	fmt.Println(err)
} 
fmt.Println("Success")
```
```go
user := &User{
	Password: "123567",
}
err := collection.UpdateById(ctx, "6QVHL46WCE680ZG2Xn3X", user, UpdatingOptions{
	SkipValidation: true,
})
if err != nil {
	fmt.Println(err)
} 
fmt.Println("Success")
```
```go
user := &User{
	Address:  &Address{
		Line1: "1 Main Road",
		City:  "New York",
	}
}
err := collection.UpdateById(ctx, "6QVHL46WCE680ZG2Xn3X", user, UpdatingOptions{
	MergeFields: []FieldPath{[]string{"address", "Line1"}},
})
if err != nil {
	fmt.Println(err)
} 
fmt.Println("Success") // only the address.Line1 field will be updated
```
- `FindById` - A method which gets the Firestore document with the specified ID. 
	- *Expects*:
		- ctx: A context.
		- id: A `string` containing the specified ID.
	- *Returns*:
		- doc: Returns the document with type `T` (the type used when initiating the collection instance).
		- error: An `error` in case something goes wrong during interaction with Firestore.
```go
user, err := collection.FindById(ctx, "6QVHL46WCE680ZG2Xn3X")
if err != nil {
	fmt.Println(err)
} 
fmt.Println(user) // {Bobby Donev hello@bobbydonev.com asdasdkjahdks 26 0xc0001d05a0}}
```
- `DeleteById` - A method which deletes the Firestore document with the specified ID. 
	- *Expects*:
		- ctx: A context.
		- id: A `string` containing the specified ID.
	- *Returns*:
		- error: An `error` in case something goes wrong during interaction with Firestore.
	- If the document does not exist, it does nothing and `error` is `nil`.
```go
err := collection.DeleteById(ctx, "6QVHL46WCE680ZG2Xn3X")
if err != nil {
	fmt.Println(err)
} 
fmt.Println("Success")
```
- `Validate` - A method which validates passed in data. 
	- *Expects*:
		- data: A `pointer` of a `struct` with populated fields which will be validated.
		- options *(optional)*: An instance of `ValidationOptions` with two properties. 
			- SkipRequired: A `bool` which when `false`, means the `required` tag will not be ignored (i.e. the `required` check is not skipped). Default value is `true`.
			- SkipValidation: A `bool` which when `true`, means all validation tags will be ingored (the `name` and `omitempty` tags will be acknowledged). Default is `false`.
			- AllowEmptyFields: An optional `slice` of type `FieldPath`, which is used to specify which fields can ignore the `omitempty` and `omitemptyupdate` tags. This can be useful when a field must be set to its zero value only on certain method calls. If left empty, all fields will honour the two tags.
	- *Returns*:
		- error: An `error` in case something goes wrong during validation.
	- ***Important***: 
		- If neither `omitempty`, nor `omitemptyupdate` tags have been used, non-specified field values in the passed in data will be set to Go's default values. 
```go
user := &User{
	Email: "HELLO@BOBBYDONEV.COM",
}
err := collection.Validate(user)
if err != nil {
	fmt.Println(err)
} 
fmt.Println(user) // {hello@bobbydonev.com}
```
```go
err := collection.Validate(user, ValidationOptions{SkipRequired: false})
if err != nil {
	fmt.Println(err)
} 
fmt.Println(user) // {hello@bobbydonev.com}
```
- `Find` - A method which returns a Firevault `query` instance.
	- *Returns*: 
		- query: A new `query` instance.
```go
query := collection.Find()
```

Queries
------------
A Firevault `query` instance allows querying Firestore, by chaining various methods. The query can have multiple filters.

- `Where` - Returns a new `query` that filters the set of results. 
	- *Expects*:
		- path: A `string` which can be a single field or a dot-separated sequence of fields.
		- operator: A `string` which must be one of `==`, `!=`, `<`, `<=`, `>`, `>=`, `array-contains`, `array-contains-any`, `in` or `not-in`.
		- value: An `interface{}` used to filter out the results.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev")
```
- `OrderBy` - Returns a new `query` that specifies the order in which results are returned. 
	- *Expects*:
		- path: A `string` which can be a single field or a dot-separated sequence of fields. To order by document name, use the special field path `DocumentID`.
		- dir: A `Direction` used to specify whether results are returned in ascending or descending order.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").OrderBy("age", Asc)
```
- `Limit` - Returns a new `query` that specifies the maximum number of first results to return. 
	- *Expects*:
		- num: An `int` which indicates the max number of results to return.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").Limit(1)
```
- `LimitToLast` - Returns a new `query` that specifies the maximum number of last results to return. 
	- *Expects*:
		- num: An `int` which indicates the max number of results to return.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").LimitToLast(1)
```
- `Offset` - Returns a new `query` that specifies the number of initial results to skip. 
	- *Expects*:
		- num: An `int` which indicates the number of results to skip.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").Offset(1)
```
- `StartAt` - Returns a new `query` that specifies that results should start at the document with the given field values.
	- *Expects*:
		- path: A `string` which can be a single field or a dot-separated sequence of fields. To filter by document name, use the special field path `DocumentID`.
		- field: An `interface{}` value used to filter out results.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").StartAt(DocumentID, "6QVHL46WCE680ZG2Xn3X")
```
- `StartAfter` - Returns a new `query` that specifies that results should start just after the document with the given field values.
	- *Expects*:
		- path: A `string` which can be a single field or a dot-separated sequence of fields. To filter by document name, use the special field path `DocumentID`.
		- field: An `interface{}` value used to filter out results.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").StartAfter(DocumentID, "6QVHL46WCE680ZG2Xn3X")
```
- `EndBefore` - Returns a new `query` that specifies that results should end just before the document with the given field values.
	- *Expects*:
		- path: A `string` which can be a single field or a dot-separated sequence of fields. To filter by document name, use the special field path `DocumentID`.
		- field: An `interface{}` value used to filter out results.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").EndBefore(DocumentID, "6QVHL46WCE680ZG2Xn3X")
```
- `EndAt` - Returns a new `query` that specifies that results should end at the document with the given field values.
	- *Expects*:
		- path: A `string` which can be a single field or a dot-separated sequence of fields. To filter by document name, use the special field path `DocumentID`.
		- field: An `interface{}` value used to filter out results.
	- *Returns*:
		- A new `query` instance.
```go
newQuery := collection.Find().Where("name", "==", "Bobby Donev").EndAt(DocumentID, "6QVHL46WCE680ZG2Xn3X")
```
- `Fetch` - Returns results based on the chained methods before it.
	- *Expects*:
		- ctx: A context.
	- *Returns*:
		- results: A `slice` containing the results of type `Document[T]` (where `T` is the type used when initiating the collection instance). `Document[T]` has two properties.
			- ID: A `string` which holds the document's ID.
			- Data: The document's data of type `T`.
		- error: An `error` in case something goes wrong.
```go
users, err := collection.Find().Where("name", "==", "Bobby Donev").Fetch(ctx)
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(users) // []Document[User]
	fmt.Println(users[0].ID) // 6QVHL46WCE680ZG2Xn3X
}
```
- `Count` - Returns the number of results based on the chained methods before it.
	- *Expects*:
		- ctx: A context.
	- *Returns*:
		- count: An `int64` representing the number of documents which meet the criteria.
		- error: An `error` in case something goes wrong.
```go
count, err := collection.Find().Where("name", "==", "Bobby Donev").Count(ctx)
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(count) // 1
}
```

Custom Errors
------------
During collection methods which require validation (i.e. `Create`, `UpdateById` and `Validate`), Firevault may return an error of a `FieldError` interface, which can aid in presenting custom error messages to users. All other errors are of the usual `error` type. Available methods for `FieldError` can be found in the `field_error.go` file. 

Here is an example of parsing returned error.
```go
func parseError(err firevault.FieldError) {
	if err.StructField() == "Password" { // or err.Field() == "password"
		if err.Tag() == "min=6" {
			fmt.Println("Password must be at least 6 characters long.")
		} else {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(err.Error())
	}
}

id, err := collection.Create(ctx, &User{
	Name: "Bobby Donev",
	Email: "hello@bobbydonev.com",
	Password: "12345",
	Age: 26,
	Address: &Address{
		Line1: "1 High Street",
		City:  "London",
	},
})
if err != nil {
	var fErr firevault.FieldError
	if errors.As(err, &fErr) {
		parseError(fErr) // "Password must be at least 6 characters long."
	} else {
		fmt.Println(err.Error())
	}
} else {
	fmt.Println(id)
}
```

Contributing
------------
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

License
------------
[MIT](https://choosealicense.com/licenses/mit/)