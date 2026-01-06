GET http://localhost:8080/

HTTP 200
[Asserts]
jsonpath "$.message" == "Hello World"
