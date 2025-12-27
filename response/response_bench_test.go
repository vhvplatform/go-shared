package response

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkSuccess(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("correlation_id", "test-correlation-id")
		b.StartTimer()
		
		Success(c, map[string]string{"status": "ok"})
	}
}

func BenchmarkSuccessWithMeta(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	meta := NewMeta(1, 10, 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("correlation_id", "test-correlation-id")
		b.StartTimer()
		
		SuccessWithMeta(c, map[string]string{"status": "ok"}, meta)
	}
}

func BenchmarkError(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("correlation_id", "test-correlation-id")
		b.StartTimer()
		
		Error(c, 400, "BAD_REQUEST", "Invalid input")
	}
}

func BenchmarkNewMeta(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMeta(1, 10, 100)
	}
}
