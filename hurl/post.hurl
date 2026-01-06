POST http://localhost:3000/data
Content-Type: application/json

{
  "name": "test",
  "value": 123
}

HTTP 200
[Asserts]
jsonpath "$.received.name" == "test"
jsonpath "$.received.value" == 123
