package services

import "fmt"

func ShowQuery(q Querier) {
	fmt.Println(q.String())
}

func ShowStorage(s MetricPrinter) {
	fmt.Println(s.String())
}
