# errors

## Usage

**Create New error**

```
import "gitscm.cisco.com/mcmp/errors"

func main() {
  err := errors.NewDomainError(errors.ErrNotFound, errors.Tenant, "tenantID: e3bff506-7b3f-45dd-914c-c2f37ca5d522")
}
```

**Convert Exception to component Error**

```
import (
  "fmt"

  "gitscm.cisco.com/mcmp/errors"

  "gitscm.cisco.com/mcmp/tenantmgr/gen/models"
)

func main() {
  err := errors.NewDomainError(errors.ErrNotFound, errors.Tenant, "tenantID: e3bff506-7b3f-45dd-914c-c2f37ca5d522")
  errm := new(models.Error)
  if cverr := errors.Convert(err, errm); cverr != nil {
    fmt.Println(cverr)
  }
}
```

**Case statement to determine type of error**

```
import "gitscm.cisco.com/mcmp/errors"

func main() {
  err := errors.NewDomainError(errors.ErrNotFound, errors.Tenant, "tenantID: e3bff506-7b3f-45dd-914c-c2f37ca5d522")
  if err != nil {
    switch e := err.(type) {
    case Exception:
      // handle Exception types
      switch e.Type() {
      case errors.ErrNotFound:
        // return 404
      case errors.ErrExists:
        // return 409
      }
    }
    // return 500
  }
  // return 200
}
```

**Check for specific type of error**

```
import "gitscm.cisco.com/mcmp/errors"

func main() {
  err := errors.NewDomainError(errors.ErrNotFound, errors.Tenant, "tenantID: e3bff506-7b3f-45dd-914c-c2f37ca5d522")
  if errors.IsType(errors.ErrNotFound, err) {
    return err
  }
  // continue logic
}
```

**Configure error handler for OpenAPI**

```
func main() {

  logger := iam.GetLogger()

  swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
  if err != nil {
    logger.Fatalln(err)
  }

  api := ops.NewIamAPI(swaggerSpec)
  api.Logger = logger.Infof
  api.ServeError = errors.DefaultHandler.ServeError

  // continue constructing main...
}
```

## Add Context and StackTrace to "other" errors

Each library imported returns will at some point return an error. Most of the time the only action taken is simply to check if error is not `nil` and return the error.
Often this leads to a strange or unrecognized error showing up in the logs that takes time to track down the source of the error.

To help provide additional context and a stack trace, the errors can be wrapped using `github.com/pkg/errors`.

**Example**

```
_, err := ioutil.ReadAll(r)
if err != nil {
  return errors.Wrap(err, "read failed")
}
```

**Logging Example**

```
func read() error {
  _, err := ioutil.ReadAll(r)
  if err != nil {
    return errors.Wrap(err, "read failed")
  }
}

func main() {
  if err := read(); err != nil {
    log.Errorf("%+v", err)  // logs the error message and the stack trace for the error
  }
}
```

## Contributions

Contributing guidelines are in [CONTRIBUTING.md](CONTRIBUTING.md).

Help Wanted with [Outstanding Tasks](TODO.md)
