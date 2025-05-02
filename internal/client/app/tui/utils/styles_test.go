package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

func TestNavigateFooter(t *testing.T) {
	type args struct {
		// no args needed
	}
	tests := []struct {
		name           string
		args           args
		wantSubstrings []string
	}{
		{
			name: "NavigateFooter contains instructions and styles",
			args: args{},
			wantSubstrings: []string{
				"Use arrow keys to navigate",
				"enter to select",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := utils.NavigateFooter()
			for _, substr := range tt.wantSubstrings {
				require.Contains(t, footer, substr)
			}
		})
	}
}

func TestAddItemsFooter(t *testing.T) {
	type args struct{}
	tests := []struct {
		name           string
		args           args
		wantSubstrings []string
	}{
		{
			name: "AddItemsFooter contains instructions",
			args: args{},
			wantSubstrings: []string{
				"Press Enter to save",
				"CTRL+Q to cancel",
				"Backspace to delete",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := utils.AddItemsFooter()
			for _, substr := range tt.wantSubstrings {
				require.Contains(t, footer, substr)
			}
		})
	}
}

func TestListItemsFooter(t *testing.T) {
	type args struct{}
	tests := []struct {
		name           string
		args           args
		wantSubstrings []string
		want           string
	}{
		{
			name: "ListItemsFooter contains instructions",
			args: args{},
			wantSubstrings: []string{
				"Use arrow keys to navigate",
				"E to edit",
				"D to delete",
				"Enter to select",
				"CTRL+Q to cancel",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ListItemsFooter()
			for _, substr := range tt.wantSubstrings {
				require.Contains(t, result, substr)
			}
		})
	}
}

func TestItemDataFooter(t *testing.T) {
	type args struct{}
	tests := []struct {
		name           string
		args           args
		wantSubstrings []string
	}{
		{
			name: "ItemDataFooter contains return instruction",
			args: args{},
			wantSubstrings: []string{
				"Press CTRL+Q to return",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ItemDataFooter()
			for _, substr := range tt.wantSubstrings {
				require.Contains(t, result, substr)
			}
		})
	}
}

func TestBinaryItemDataFooter(t *testing.T) {
	type args struct{}
	tests := []struct {
		name           string
		args           args
		wantSubstrings []string
		want           string
	}{
		{
			name: "BinaryItemDataFooter contains download instruction",
			args: args{},
			wantSubstrings: []string{
				"D to download file",
				"CTRL+Q to return",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := utils.BinaryItemDataFooter()
			for _, substr := range tt.wantSubstrings {
				require.Contains(t, footer, substr)
			}
		})
	}
}

func TestAuthFooter(t *testing.T) {
	type args struct{}
	tests := []struct {
		name           string
		args           args
		wantSubstrings []string
	}{
		{
			name: "AuthFooter contains instructions",
			args: args{},
			wantSubstrings: []string{
				"Press Tab to switch fields",
				"Enter to submit",
				"CTRL+Q to exit",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.AuthFooter()
			for _, substr := range tt.wantSubstrings {
				require.Contains(t, result, substr)
			}
		})
	}
}
