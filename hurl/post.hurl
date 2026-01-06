POST http://localhost:8080/data
Content-Type: application/json

{
  "name": "test",
  "value": 123
}

HTTP 200
[Asserts]
jsonpath "$.received.name" == "test"
jsonpath "$.received.value" == 123
