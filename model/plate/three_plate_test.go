package plate

import (
	"reflect"
	"testing"
)

func Test_newThreePlate(t *testing.T) {
	type args struct {
		a int
		b int
		c int
	}
	tests := []struct {
		name string
		args args
		want *ThreePlate
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newThreePlate(tt.args.a, tt.args.b, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newThreePlate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThreePlate_hash(t *testing.T) {
	type fields struct {
		Compare Compare
		Plate   []int
		Level   ThreePlateLevel
		Value   int
	}
	tests := []struct {
		name   string
		fields fields
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ThreePlate{
				Compare: tt.fields.Compare,
				Plate:   tt.fields.Plate,
				Level:   tt.fields.Level,
				Value:   tt.fields.Value,
			}
			p.hash()
		})
	}
}

func TestThreePlateSet_Get(t *testing.T) {
	type fields struct {
		Generate Generate
		Plates   chan int
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ThreePlateSet{
				Generate: tt.fields.Generate,
				Plates:   tt.fields.Plates,
			}
			if got := s.Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ThreePlateSet.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewThreePlateSet(t *testing.T) {
	tests := []struct {
		name string
		want *ThreePlateSet
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			want: nil,
		},
		{
			name: "test2",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewThreePlateSet(); !reflect.DeepEqual(got, tt.want) {
				//t.Errorf("NewThreePlateSet() = %v, want %v", got, tt.want)

				if tp,ok:=got.Get().(*ThreePlate);ok{
					t.Log(tp.Plate)
				}else{
					t.Error("threeplate error")
				}
			}
		})
	}
}

func TestThreePlate_Less(t *testing.T) {
	type fields struct {
		Compare Compare
		Plate   []int
		Level   ThreePlateLevel
		Value   int
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ThreePlate{
				Compare: tt.fields.Compare,
				Plate:   tt.fields.Plate,
				Level:   tt.fields.Level,
				Value:   tt.fields.Value,
			}
			if got := p.Less(tt.args.v); got != tt.want {
				t.Errorf("ThreePlate.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}
