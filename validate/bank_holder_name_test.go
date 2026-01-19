package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBankHolderName(t *testing.T) {
	type args struct {
		inputName           string
		targetName          string
		levenshteinDistance int
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Test pass holder name",
			args: args{
				inputName:  "NGUYEN VAN NU",
				targetName: "Nguyễn Văn Nữ",
			},
			wantErr: nil,
		},
		{
			name: "Test holder name is empty",
			args: args{
				inputName:  " ",
				targetName: "Nguyễn Văn Nữ",
			},
			wantErr: ErrInputNameEmpty,
		},
		{
			name: "Test holder name is not have white space add the middle",
			args: args{
				inputName:  "NGUYENVANU",
				targetName: "Nguyễn Văn Nữ",
			},
			wantErr: ErrInputNameEmpty,
		},
		{
			name: "Test ekyc full name is empty",
			args: args{
				inputName:  "",
				targetName: "NGUYEN VAN NU",
			},
			wantErr: ErrInputNameEmpty,
		},
		{
			name: "Test holder name is grander than target name",
			args: args{
				inputName:  "NGUYỄN VĂN NAM",
				targetName: "NGUYEN VAN NU",
			},
			wantErr: ErrInvalidInputNameLength,
		},
		{
			name: "Test ekyc full name not have white space",
			args: args{
				inputName:  "NguyễnVănNam",
				targetName: "NGUYEN VAN NU",
			},
			wantErr: ErrInputNameEmpty,
		},
		{
			name: "Test Unicode especially case",
			args: args{
				inputName:  "PHẠM ĐỨC HẢI",
				targetName: "PHAM DUC HAI",
			},
			wantErr: nil,
		},
		{
			name: "Test unicode from typing in editor",
			args: args{
				inputName:  "PHẠM ĐỨC HẢI",
				targetName: "PHAM DUC HAI",
			},
			wantErr: nil,
		},
		{
			name: "Test check Levenshtein Distance algorithm failed: (Config max Levenshtein Distance = 3)",
			args: args{
				inputName:           "HA DU HA",
				targetName:          "PHẠM ĐỨC HẢI",
				levenshteinDistance: 3,
			},
			wantErr: ErrInvalidLevenshtein,
		},
		{
			name: "Invalid target name",
			args: args{
				inputName:  "NGUYEN VAN NU",
				targetName: " ",
			},
			wantErr: ErrTargetNameEmpty,
		},
		{
			name: "Invalid first name",
			args: args{
				inputName:           "NGUYEN VAN NU",
				targetName:          "NGUYEN VAN NAM",
				levenshteinDistance: 3,
			},
			wantErr: ErrInvalidFirstName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BankHolderName(tt.args.inputName, tt.args.targetName, tt.args.levenshteinDistance)
			assert.Equalf(t, err, tt.wantErr, "BankHolderName() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
