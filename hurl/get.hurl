GET http://localhost:3000/

HTTP 200
[Asserts]
jsonpath "$.message" == "Hello World"
