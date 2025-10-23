package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Nested struct {
		User User `validate:"nested"`
	}
)

//nolint:funlen
func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		wantErr     bool
		errContains []string // подстроки в ошибке (для ValidationErrors)
		progErr     error    // если ожидается программная ошибка
	}{
		{
			name: "valid user",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789abc", // 36 chars
				Age:    25,
				Email:  "user@example.com",
				Role:   "admin",
				Phones: []string{"12345678901", "11111111111"},
			},
			wantErr: false,
		},
		{
			name: "invalid ID length",
			in: User{
				ID:  "short",
				Age: 25,
			},
			wantErr:     true,
			errContains: []string{"ID: length must be 36"},
		},
		{
			name: "age below min",
			in: User{
				ID:  "12345678-1234-1234-1234-123456789abc",
				Age: 10,
			},
			wantErr:     true,
			errContains: []string{"Age: value 10 must be greater than or equal to 18"},
		},
		{
			name: "invalid email",
			in: User{
				ID:    "12345678-1234-1234-1234-123456789abc",
				Age:   25,
				Email: "invalid-email",
			},
			wantErr:     true,
			errContains: []string{"Email: regexp must match"},
		},
		{
			name: "invalid role",
			in: User{
				ID:   "12345678-1234-1234-1234-123456789abc",
				Age:  25,
				Role: "guest",
			},
			wantErr:     true,
			errContains: []string{"Role: value \"guest\" not in allowed list"},
		},
		{
			name: "invalid phone length",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789abc",
				Age:    25,
				Phones: []string{"123"},
			},
			wantErr:     true,
			errContains: []string{"length must be 11"},
		},
		{
			name: "valid app",
			in: App{
				Version: "1.2.3",
			},
			wantErr: false,
		},
		{
			name: "invalid app version length",
			in: App{
				Version: "1.2",
			},
			wantErr:     true,
			errContains: []string{"Version: length must be 5"},
		},
		{
			name: "valid response",
			in: Response{
				Code: 200,
			},
			wantErr: false,
		},
		{
			name: "invalid response code",
			in: Response{
				Code: 400,
			},
			wantErr:     true,
			errContains: []string{"Code: value 400 not in allowed set {200,404,500}"},
		},
		{
			name:    "not a struct",
			in:      "not a struct",
			wantErr: true,
			progErr: ErrInvalidType,
		},
		{
			name:    "empty struct",
			in:      struct{}{},
			wantErr: true,
			progErr: ErrEmptyStruct,
		},
		{
			name: "nested valid",
			in: Nested{
				User: User{
					ID:    "12345678-1234-1234-1234-123456789abc",
					Age:   25,
					Role:  "admin",
					Email: "user@example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "nested invalid",
			in: Nested{
				User: User{
					ID:  "short",
					Age: 25,
				},
			},
			wantErr:     true,
			errContains: []string{"User.ID: length must be 36"},
		},
		{
			name: "unsupported validator",
			in: struct {
				Value string `validate:"unknown:5"`
			}{},
			wantErr: true,
			progErr: ErrTagUnsupportedValidator,
		},
		{
			name: "empty validator",
			in: struct {
				Value string `validate:""`
			}{},
			wantErr: false, // no tag → ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.in)

			if tt.progErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.progErr), "expected program error %v, got %v", tt.progErr, err)
				return
			}

			if tt.wantErr {
				require.Error(t, err)
				var valErrs ValidationErrors
				require.True(t, errors.As(err, &valErrs), "expected ValidationErrors, got %T", err)
				errStr := err.Error()
				for _, substr := range tt.errContains {
					assert.Contains(t, errStr, substr, "error should contain %q", substr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
