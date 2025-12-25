# SMS Service Package

Package `sms` cung cấp abstraction layer để gửi SMS với hỗ trợ nhiều provider khác nhau.

## Providers Hỗ Trợ

- ✅ Twilio API
- ✅ AWS Simple Notification Service (SNS)
- ✅ Vonage/Nexmo API
- ✅ MessageBird API

## Tính Năng

- Gửi SMS đơn giản
- Gửi hàng loạt (bulk sending)
- Theo dõi trạng thái tin nhắn
- Hỗ trợ Unicode
- Tính toán segments và chi phí
- Validation số điện thoại

## Sử Dụng Cơ Bản

```go
import "github.com/vhvcorp/go-shared/sms"

// Tạo client với Twilio
client, err := sms.NewClient(sms.Config{
    Provider: sms.ProviderTwilio,
    From:     "+1234567890",
    Options: map[string]string{
        "account_sid": "your-account-sid",
        "auth_token": "your-auth-token",
    },
})

// Gửi SMS
msg := &sms.Message{
    From: "+1234567890",
    To:   []string{"+0987654321"},
    Body: "Hello from Go!",
}

result, err := client.Send(ctx, msg)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("SMS sent with ID: %s, Status: %s\n", 
    result.MessageID, result.Status)
```

## Tính Toán Segments

```go
msg := &sms.Message{
    Body: "Very long message...", // 200 characters
}

segments := msg.CalculateSegments()
fmt.Printf("Message will use %d segments\n", segments)

// Ước tính chi phí
cost := msg.EstimateCost(0.05) // $0.05 per segment
fmt.Printf("Estimated cost: $%.2f\n", cost)
```

## Kiểm Tra Trạng Thái

```go
// Sau khi gửi
result, _ := client.Send(ctx, msg)

// Kiểm tra trạng thái
time.Sleep(5 * time.Second)
status, err := client.GetStatus(ctx, result.MessageID)
fmt.Printf("Current status: %s\n", status.Status)
```

## Gửi Hàng Loạt

```go
messages := []*sms.Message{
    {To: []string{"+1111111111"}, Body: "Message 1"},
    {To: []string{"+2222222222"}, Body: "Message 2"},
}

results, err := client.SendBulk(ctx, messages)
```

## Unicode Support

```go
msg := &sms.Message{
    From:    "+1234567890",
    To:      []string{"+0987654321"},
    Body:    "Xin chào! 你好！",
    Unicode: true, // Bật hỗ trợ Unicode
}

client.Send(ctx, msg)
```

## Status

🚧 **In Development** - Hiện tại package đã có interfaces và structures sẵn sàng, implementations cho các providers đang được phát triển.
