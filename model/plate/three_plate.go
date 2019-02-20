package plate

import (
	"math/rand"
	"sort"
	"time"
)

type Compare interface {
	Less(interface{}) bool
}
type Generate interface {
	Get()interface{}
}

type ThreePlateLevel int

const (
	// 梅花 100
	// 方块 200
	// 红心 300
	// 黑桃 400
	plate_max=54

	plate_club = 100
	plate_diamond=200
	plate_heart=300
	plate_spade=400

	plate_Joker2 = 16
	plate_Joker1 = 15
	plate_A      = 14
	plate_K      = 13
	plate_Q      = 12
	plate_J      = 11
	plate_10     = 10
	plate_9      = 9
	plate_8      = 8
	plate_7      = 7
	plate_6      = 6
	plate_5      = 5
	plate_4      = 4
	plate_3      = 3
	plate_2      = 2

	TP_Value_Single     = 1
	TP_Value_Double     = 1000
	TP_Value_Order      = 100000
	TP_Value_Equal      = 1000000
	TP_Value_OrderEqual = 1000000000
	TP_Value_Three      = 10000000000

	/*

		三同占用一位(数字)
		顺金占用一位(最大的数字)
		金花占用三位(三个不同的数字)
		顺子占用一位(最大的数字)
		对子占用两位(对子的数字和单牌的数字)
		单排占用三位(三个不同的数字)

	*/
	TP_Level_Single ThreePlateLevel = iota
	TP_Level_Double
	TP_Level_Order
	TP_Level_Equal
	TP_Level_OrderEqual
	TP_Level_Three
)
func newThreePlate(a, b, c int) *ThreePlate {
	res := &ThreePlate{
		Plate: []int{a, b, c},
	}
	res.hash()
	return res
}
func (p *ThreePlate) hash() {

	//排序
	sort.Slice(p.Plate, func(i, j int) bool {
		return p.Plate[i] < p.Plate[j]
	})
	//数字
	a0 := p.Plate[0] % 100
	a1 := p.Plate[1] % 100
	a2 := p.Plate[2] % 100
	//花色
	b0 := p.Plate[0] / 100
	b1 := p.Plate[0] / 100
	b2 := p.Plate[0] / 100

	is_single := func() int {
		return a2*100 + a1*10 + a0
	}
	is_double := func() int {
		//对比单牌大
		if a2 == a1 {
			return a2*10 + a0
		}
		//对比单牌小
		if a1 == a0 {
			return a1*10 + a2
		}
		return 0
	}
	is_three := func() int {
		//三个数字一样的
		if a0 == a1 && a1 == a2 {
			return a2
		}
		return 0
	}
	is_equal := func() int {
		//花色一样的
		if b0 == b1 && b1 == b2 {
			return is_single()
		}
		return 0
	}
	is_order := func() int {
		if a2 == plate_A {
			//123
			if a0 == plate_2 && a1 == plate_3 {
				return a1
			}
			//QKA
			if a0 == plate_Q && a1 == plate_K {
				return a2
			}
			return 0
		} else {
			//234 JQK
			if a0+1 != a1 || a1+1 != a2 {
				return a2
			}
			return a2
		}
	}

	three := is_three()
	if three > 0 {
		p.Value = three * TP_Value_Three
		p.Level = TP_Level_Three
		return
	}
	order := is_order()
	equal := is_equal()
	if equal > 0 {
		if order > 0 {
			p.Value = order * TP_Value_OrderEqual
			p.Level = TP_Level_OrderEqual

			return
		} else {
			p.Value = equal * TP_Value_Equal
			p.Level = TP_Level_Equal
			return
		}
	}
	if order > 0 {
		p.Value = order * TP_Value_Order
		p.Level = TP_Level_Order
		return
	}
	double := is_double()
	if double > 0 {
		p.Value = double * TP_Value_Double
		p.Level = TP_Level_Double
		return
	}
	p.Value = is_single()
	p.Level = TP_Level_Single
	return
}

type ThreePlate struct {
	Compare
	Plate []int
	Level ThreePlateLevel
	Value int
}
type ThreePlateSet struct {
	Generate
	Plates chan int
}
func (s *ThreePlateSet)Get()interface{}{
	if len(s.Plates)<3 {
		return nil
	}
	a:=<-s.Plates
	b:=<-s.Plates
	c:=<-s.Plates
	tp:=newThreePlate(a,b,c)
	return tp
}

func NewThreePlateSet()*ThreePlateSet{
	set:=ThreePlateSet{
		Plates:make(chan int,plate_max),
	}
	plateList:=[]int{}

	for i:=plate_club;i<=plate_spade;i+=100{
		for j:=plate_2;j<=plate_A;j++{
			plateList=append(plateList,i+j)
		}
	}
	rand.Seed(time.Now().UnixNano())
	for i:=plate_max-2;i>0;i--{
		index:=rand.Intn(i)
		set.Plates<-plateList[index]
		plateList=append(plateList[:index],plateList[index+1:]...)
	}
	return &set
}


func (p *ThreePlate) Less(v interface{}) bool {
	other, ok := v.(*ThreePlate)
	if !ok {
		return false
	}
	return p.Value <= other.Value
}
