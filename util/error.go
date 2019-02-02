package util

import (
	"fmt"
)

func ToString(args ...interface{})string{
	str:=""
	for i,v:=range args {
		str+=fmt.Sprintf("%v", v)
		if i!=len(args)-1{
			str+=" "
		}
	}
	return str
}
