package providers

import (
	"chat_app_backend/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinioProviderGetFileURL(t *testing.T) {
	tests := []struct {
		name  string
		minio config.MinIOConfig
		path  string
		want  string
	}{
		{
			name:  "MinIO 公開網址附加 bucket",
			minio: config.MinIOConfig{PublicURL: "https://images.example.com/"},
			path:  "/server/a.png",
			want:  "https://images.example.com/chat-app/server/a.png",
		},
		{
			name:  "R2 自訂網域不附加 bucket",
			minio: config.MinIOConfig{PublicURL: "https://images.example.com", PublicBaseURL: "https://chat-images.example.com/"},
			path:  "server/a.png",
			want:  "https://chat-images.example.com/server/a.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &minioProvider{cfg: &config.Config{MinIO: tt.minio}, bucket: "chat-app"}
			assert.Equal(t, tt.want, mp.GetFileURL(tt.path))
		})
	}
}
