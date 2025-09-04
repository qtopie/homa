# homa

build plugins 

```
go build -buildmode=plugin -o mock.so
```

test

```
calvin.y@Mac go-futu-api % grpcurl -plaintext -d '{"message": "Hello, Service Two!"}' localhost:1234 main.ChatService.Chat
{
  "content": "Hello, Service Two! - Response 2025-08-28T22:14:08+08:00"
}
{
  "content": "Hello, Service Two! - Response 2025-08-28T22:14:09+08:00"
}
{
  "content": "Hello, Service Two! - Response 2025-08-28T22:14:10+08:00"
}
{
  "content": "Hello, Service Two! - Response 2025-08-28T22:14:11+08:00"
}
{
  "content": "Hello, Service Two! - Response 2025-08-28T22:14:12+08:00"
}
```