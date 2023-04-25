package main

import (
	"fmt"
	podrestarterv1alpha1 "github.com/myoperator/poddeployer/pkg/apis/podDeployer/v1alpha1"
	"sort"
)

func main() {
	ex := make([]podrestarterv1alpha1.PriorityImage, 0)
	a := podrestarterv1alpha1.PriorityImage{Image: "aaa", Value: 10}
	b := podrestarterv1alpha1.PriorityImage{Image: "bbb", Value: 200}
	c := podrestarterv1alpha1.PriorityImage{Image: "ccc", Value: 1000}
	ex = append(ex, a)
	ex = append(ex, b)
	ex = append(ex, c)

	sort.SliceStable(ex, func(i, j int) bool {
		return ex[i].Value > ex[j].Value
	})

	fmt.Println(ex)

}



