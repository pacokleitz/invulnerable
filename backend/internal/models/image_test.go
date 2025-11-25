package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImage_FullName(t *testing.T) {
	tests := []struct {
		name     string
		image    Image
		expected string
	}{
		{
			name: "with registry",
			image: Image{
				Registry:   "docker.io",
				Repository: "library/nginx",
				Tag:        "latest",
			},
			expected: "docker.io/library/nginx:latest",
		},
		{
			name: "without registry",
			image: Image{
				Registry:   "",
				Repository: "nginx",
				Tag:        "alpine",
			},
			expected: "nginx:alpine",
		},
		{
			name: "with custom registry",
			image: Image{
				Registry:   "gcr.io",
				Repository: "myproject/myimage",
				Tag:        "v1.0.0",
			},
			expected: "gcr.io/myproject/myimage:v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.image.FullName()
			assert.Equal(t, tt.expected, result)
		})
	}
}
