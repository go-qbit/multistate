package multistate

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-qbit/multistate/expr"
)

type Implementation interface{}

type Action struct {
	Caption    string
	From       expr.Expression
	Set        States
	Reset      States
	OnDo       ActionDoFunc
	Availabler Availabler
}

type Cluster struct {
	Caption string
	Expr    expr.Expression
}

func NewFromStruct(s Implementation) *Multistate {
	mst := New("New")

	rtS := reflect.TypeOf(s)
	rvS := reflect.ValueOf(s)

	rvStruct := rvS
	rtStruct := rtS
	if rvStruct.Kind() == reflect.Ptr {
		rvStruct = rvStruct.Elem()
		rtStruct = rtStruct.Elem()
	}
	for i := 0; i < rvStruct.NumField(); i++ {
		ft := rtStruct.Field(i)
		if ft.Type.String() != "multistate.State" {
			continue
		}

		strBit, exists := ft.Tag.Lookup("bit")
		if !exists {
			panic(fmt.Sprintf("Missed required tag 'bit' for field '%s'", ft.Name))
		}
		bit, err := strconv.ParseUint(strBit, 10, 7)
		if err != nil {
			panic(fmt.Sprintf("Invalid 'bit' value for field '%s'", ft.Name))
		}

		caption := ft.Name
		if t, exists := ft.Tag.Lookup("caption"); exists {
			caption = t
		}

		rvStruct.Field(i).Set(reflect.ValueOf(mst.MustAddState(uint8(bit), camelCaseToSnake(ft.Name), caption)))
	}

	for i := 0; i < rtS.NumMethod(); i++ {
		mt := rtS.Method(i)

		if strings.HasPrefix(mt.Name, "Action") {
			values := rvS.Method(i).Call(nil)
			if len(values) != 1 || values[0].Type().String() != "multistate.Action" {
				panic(fmt.Sprintf("The action method %s must return the multistate.Action structure", mt.Name))
			}

			action := values[0].Interface().(Action)
			caption := mt.Name[6:]
			if action.Caption != "" {
				caption = action.Caption
			}

			if action.From == nil {
				action.From = expr.Empty()
			}

			mst.MustAddAction(camelCaseToSnake(mt.Name[6:]), caption, action.From, action.Set, action.Reset, action.OnDo, action.Availabler)
		} else if mt.Name == "OnDoAction" {
			cb, ok := rvS.Method(i).Interface().(func(context.Context, Entity, uint64, uint64, string, ...interface{}) error)
			if !ok {
				panic(fmt.Sprintf("OnDoAction must fit OnDoCallback type "))
			}
			mst.SetOnDoCallback(cb)
		} else if mt.Name == "Clusters" {
			values := rvS.Method(i).Call(nil)
			clusters, ok := values[0].Interface().([]Cluster)
			if !ok {
				panic(fmt.Sprintf("The Clusters method must return the []multistate.Cluster type"))
			}
			for _, c := range clusters {
				mst.AddCluster(c.Caption, c.Expr)
			}
		}
	}

	mst.MustCompile()

	return mst
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelCaseToSnake(s string) string {
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
