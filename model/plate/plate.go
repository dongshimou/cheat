package plate

import (
	"sort"
)

type Compare interface {
	Less(interface{}) bool
}

type ThreePlateLevel int
const (
	plate_A  = 14
	plate_K  = 13
	plate_Q  = 12
	plate_J  = 11
	plate_10 = 10
	plate_9  = 9
	plate_8  = 8
	plate_7  = 7
	plate_6  = 6
	plate_5  = 5
	plate_4  = 4
	plate_3  = 3
	plate_2  = 2

	plate_Single    = 1
	plate_Double    = 1000
	plate_Order      = 100000
	plate_Equal     = 1000000
	plate_OrderEqual = 1000000000
	plate_Three      = 10000000000

	PlateSingle ThreePlateLevel = iota
	PlateDouble
	PlateOrder
	PlateEqual
	PlateOrderEqual
	PlateThree
)

func NewPlate(a,b,c int)*ThreePlate {
	res:=&ThreePlate{
		Plate:[3]int{a,b,c},
	}
	res.hash()
	return res
}
func (p *ThreePlate) hash(){

	//排序
	sort.Slice(p.Plate, func(i, j int) bool {
		return p.Plate[i]< p.Plate[j]
	})
	//大小
	a0:= p.Plate[0]%100
	a1:= p.Plate[1]%100
	a2:= p.Plate[2]%100
	//花色
	b0:= p.Plate[0]/100
	b1:= p.Plate[0]/100
	b2:= p.Plate[0]/100

	is_single:= func() int{
		return a2*100+a1*10+a0
	}
	is_double:= func() int {
		if a2 == a1 {
			return a2*10 + a0
		}
		if a1 == a0 {
			return a1*10 + a2
		}
		return 0
	}
	is_three:= func() int{
		if a0==a1&&a1==a2{
			return a2
		}
		return 0
	}
	is_equal:= func() int{
		if b0==b1&&b1==b2{
			return is_single()
		}
		return 0
	}
	is_order:= func() int{
		if a2==plate_A{
			if a0==plate_2&&a1==plate_3{
				return a1
			}
			if a0==plate_Q&&a1==plate_K{
				return a2
			}
			return 0
		}else{
			if a0+1!=a1||a1+1!=a2{
				return a2
			}
			return a2
		}
	}

	three:=is_three()
	if three>0{
		p.Value=three*plate_Three
		p.Level=PlateThree
		return
	}
	order:=is_order()
	equal:=is_equal()
	if equal>0{
		if order>0{
			p.Value=order*plate_OrderEqual
			p.Level=PlateOrderEqual

			return
		}else{
			p.Value=equal*plate_Equal
			p.Level=PlateEqual
			return
		}
	}
	if order>0{
		p.Value=order*plate_Order
		p.Level=PlateOrder
		return
	}
	double:=is_double()
	if double>0{
		p.Value=double*plate_Double
		p.Level=PlateDouble
		return
	}
	p.Value=is_single()
	p.Level=PlateSingle
	return
}

type ThreePlate struct {
	Compare
	Plate [3]int
	// 梅花 2~A 102 ~ 114
	// 方块 2~A 202 ~ 214
	// 红心 2~A 302 ~ 314
	// 黑桃 2~A 402 ~ 414
	// Joker1 = 15
	// Joker2 = 16
	Level ThreePlateLevel
	Value int
}

func (p *ThreePlate) Less(v interface{}) bool {
	other, ok := v.(*ThreePlate)
	if !ok {
		return false
	}
	return p.Value<other.Value
}