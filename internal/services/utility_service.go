package services

import "fmt"

func ShowQuery(q Querier) {
	fmt.Println(q.String())
}

func ShowStorage(s Storager) {
	fmt.Println(s.String())
}
