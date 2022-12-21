/*
Package errors provides a common error framework to return consistent errors with a common format.

For OpenAPI Errors when initializing the APIs and binding to REST API server set `ServerError`

	api.ServeError = errors.DefaultHandler.ServeError


Example of creating a domain error

	errors.NewDomainError(errors.ErrNotFound, errors.Default, "tenant " + tenantID)


Converting a potential domain error into a domain error

	errors.RestoreDomain(err)


Check if domain error is a specific type of error

	errors.IsType(errors.ErrNotFound, err)

*/
package errors
