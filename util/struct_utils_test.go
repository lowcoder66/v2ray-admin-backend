package util

import (
	"fmt"
	"testing"
)

func TestCopyFields(t *testing.T) {
	type (
		TypeA struct {
			Age  int
			Name string
			B    bool
		}
		TypeB struct {
			Age     int
			Name    string
			B       bool
			Address string
		}
	)

	a := &TypeA{20, "张三", true}
	b := &TypeB{30, "李四", false, "成都"}
	bb := &TypeB{35, "王五", false, "北京"}

	// a -> b
	fmt.Println("before: ", a, b)
	CopyFields(a, b)
	fmt.Println("after: ", a, b)

	// bb -> a
	fmt.Println("before: ", bb, a)
	CopyFields(bb, a)
	fmt.Println("after: ", bb, a)

	// ignoreFields b -> bb
	fmt.Println("before ignoreFields (age, address): ", b, bb)
	CopyFields(b, bb, "Age", "Address")
	fmt.Println("after: ", b, bb)
}
