# Errors

Standardizing HTTP error responses

## Error

| Field       | Type        | Description                                                     |
| ----------- | ----------- | --------------------------------------------------------------- |
| Code        | int         | The HTTP response code that would be associated with this error | 
| Message     | string      | The error message to convey                                     |
| Details     | string      | (Optional) additional details relevant to the error             |

## ErrorResponse

| Field       | Type        | Description                                                     |
| ----------- | ----------- | --------------------------------------------------------------- |
| Code        | int         | The highest HTTP response code found among the errors           | 
| Errors      | []Error     | The errors included in the response                             |
| RequestID   | string      | (Optional) the ID associated with the request                   |


Errors can be be treated as a collection. For a collection of errors the highest Code for the available errors is used, defaulting to 500.

If no errors matching the Error type are included a default error response will be used.

When writing errors to a HTTP response the result will be JSON that looks like
``` json
{
	"errors": [
		{
		"message": "not authorized"
		}, 
		{
			"message": "resource not found",
			"details": "test-resource"
		}
	]
}
```


