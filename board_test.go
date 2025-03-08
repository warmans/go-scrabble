package scrabble

import (
	"reflect"
	"testing"
)

func TestBoard_getNextVerticalCellId(t *testing.T) {
	type args struct {
		cellID int64
		offset int
	}
	tests := []struct {
		name string
		b    Board
		args args
		want int64
	}{
		{
			name: "valid vertical cell",
			b:    NewBoard(15),
			args: args{1, 1},
			want: 16,
		},
		{
			name: "valid vertical cell negative offset",
			b:    NewBoard(15),
			args: args{113, -2},
			want: 83,
		},
		{
			name: "returns -1 if next cell exceeds board boundary",
			b:    NewBoard(15),
			args: args{218, 1},
			want: -1,
		},
		{
			name: "returns -1 if previous cell is before first cell",
			b:    NewBoard(15),
			args: args{23, -5},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.getNextVerticalCellId(tt.args.cellID, tt.args.offset); got != tt.want {
				t.Errorf("getNextVerticalCellId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoard_getNextHorizontalCellId(t *testing.T) {
	type args struct {
		cellID int64
		offset int
	}
	tests := []struct {
		name string
		b    Board
		args args
		want int64
	}{
		{
			name: "valid next letter",
			b:    NewBoard(15),
			args: args{1, 1},
			want: 2,
		},
		{
			name: "valid next letter with negative offset",
			b:    NewBoard(15),
			args: args{2, -1},
			want: 1,
		},
		{
			name: "returns -1 if rows spanned",
			b:    NewBoard(15),
			args: args{15, 1},
			want: -1,
		},
		{
			name: "returns -1 if rows spanned",
			b:    NewBoard(15),
			args: args{31, -1},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.getNextHorizontalCellId(tt.args.cellID, tt.args.offset); got != tt.want {
				t.Errorf("getNextHorizontalCellId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoard_nonEmptyNeighbouringCells(t *testing.T) {
	type args struct {
		cellID int64
	}
	tests := []struct {
		name string
		b    Board
		args args
		want Neighbours
	}{
		{
			name: "no neighbours",
			b:    NewBoard(15),
			args: args{cellID: 113},
			want: Neighbours{
				L: false,
				R: false,
				A: false,
				B: false,
			},
		},
		{
			name: "R neighbours",
			b: NewBoard(
				15,
				InitialWord{
					Word: "foo", Placement: Placement{
						CellId:    99,
						Direction: Down,
					},
				},
			),
			args: args{cellID: 113},
			want: Neighbours{
				L: false,
				R: true,
				A: false,
				B: false,
			},
		},
		{
			name: "L neighbours",
			b: NewBoard(
				15,
				InitialWord{
					Word: "foo", Placement: Placement{
						CellId:    97,
						Direction: Down,
					},
				},
			),
			args: args{cellID: 113},
			want: Neighbours{
				L: true,
				R: false,
				A: false,
				B: false,
			},
		},
		{
			name: "A neighbours",
			b: NewBoard(
				15,
				InitialWord{
					Word: "foo", Placement: Placement{
						CellId:    97,
						Direction: Across,
					},
				},
			),
			args: args{cellID: 113},
			want: Neighbours{
				L: false,
				R: false,
				A: true,
				B: false,
			},
		},
		{
			name: "B neighbours",
			b: NewBoard(
				15,
				InitialWord{
					Word: "foo", Placement: Placement{
						CellId:    127,
						Direction: Across,
					},
				},
			),
			args: args{cellID: 113},
			want: Neighbours{
				L: false,
				R: false,
				A: false,
				B: true,
			},
		},
		{
			name: "neighbours that span X board boundary are ignored",
			b: NewBoard(
				15,
				InitialWord{
					Word: "foo", Placement: Placement{
						CellId:    75,
						Direction: Across,
					},
				},
			),
			args: args{cellID: 113},
			want: Neighbours{
				L: false,
				R: false,
				A: false,
				B: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.nonEmptyNeighbouringCells(tt.args.cellID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nonEmptyNeighbouringCells() = %v, want %v", got, tt.want)
			}
		})
	}
}
