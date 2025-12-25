# Email Service Package

Package `email` cung cấp abstraction layer để gửi email với hỗ trợ nhiều provider khác nhau.

## Providers Hỗ Trợ

- ✅ SMTP (standard protocol)
- ✅ SendGrid API
- ✅ AWS Simple Email Service (SES)
- ✅ Mailgun API

## Tính Năng

- Gửi email đơn giản với HTML hoặc plain text
- Gửi hàng loạt (bulk sending)
- File đính kèm (attachments)
- Custom headers
- Priority levels
- Validation địa chỉ email

## Sử Dụng Cơ Bản

```go
import "github.com/vhvcorp/go-shared/email"

// Tạo client với SMTP
client, err := email.NewClient(email.Config{
    Provider: email.ProviderSMTP,
    From:     "noreply@example.com",
    Options: map[string]string{
        "host": "smtp.gmail.com",
        "port": "587",
        "username": "your-email@gmail.com",
        "password": "your-app-password",
    },
})

// Tạo và gửi message
msg := &email.Message{
    From:    "noreply@example.com",
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    Body:    "<h1>Welcome to our service</h1>",
    HTML:    true,
}

result, err := client.Send(ctx, msg)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Email sent with ID: %s\n", result.MessageID)
```

## Cấu Hình Providers

### SMTP

```go
config := email.Config{
    Provider: email.ProviderSMTP,
    Options: map[string]string{
        "host": "smtp.example.com",
        "port": "587",
        "username": "user",
        "password": "pass",
        "use_tls": "true",
    },
}
```

### SendGrid

```go
config := email.Config{
    Provider: email.ProviderSendGrid,
    Options: map[string]string{
        "api_key": "your-sendgrid-api-key",
    },
}
```

### AWS SES

```go
config := email.Config{
    Provider: email.ProviderAWSSES,
    Options: map[string]string{
        "region": "us-east-1",
        "access_key_id": "your-access-key",
        "secret_access_key": "your-secret-key",
    },
}
```

## Gửi với Attachments

```go
msg := &email.Message{
    From:    "noreply@example.com",
    To:      []string{"user@example.com"},
    Subject: "Invoice",
}

// Đọc file
fileContent, _ := os.ReadFile("invoice.pdf")

// Thêm attachment
msg.AddAttachment("invoice.pdf", fileContent, "application/pdf")

client.Send(ctx, msg)
```

## Gửi Hàng Loạt

```go
messages := []*email.Message{
    {To: []string{"user1@example.com"}, Subject: "Hello 1", Body: "Content 1"},
    {To: []string{"user2@example.com"}, Subject: "Hello 2", Body: "Content 2"},
}

results, err := client.SendBulk(ctx, messages)
for _, result := range results {
    fmt.Printf("Sent: %s\n", result.MessageID)
}
```

## Status

🚧 **In Development** - Hiện tại package đã có interfaces và structures sẵn sàng, implementations cho các providers đang được phát triển.
