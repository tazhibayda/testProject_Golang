package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tazhibayda/testProject_Golang/cdek"
	"strings"
)

func main() {
	test := flag.Bool("test", false, "Sort")

	post := cdek.NewCDEK(User, Password, ApiURLTest)

	if *test {
		post = cdek.NewCDEK(User, Password, ApiURLTest)
	}

	flag.Parse()

	addrFrom := "пр. Ленинградский, д.4"
	addrTo := "ул. Блюхера, 32"
	size := cdek.Size{
		Height: 10,
		Length: 10,
		Weight: 4000,
		Width:  10,
	}

	check, address, err := post.ValidateAddress(addrFrom)
	if err != nil {
		fmt.Println(err)
	}
	if check {
		fmt.Println(address)
	}
	doCalculate(post, test)
	a, _ := post.CreateOrder(addrFrom, addrTo, size, 7)
	fmt.Println(a)
	status, err := post.GetStatus(a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(status)

	fmt.Println(post.ValidateAddress("Заводоуковск, ул. Теплякова, 1В"))
}

func doCalculate(post *cdek.API, test *bool) {
	tariffDescription := flag.String("tariff-description", " ", "Sort")
	tariffName := flag.String("tariff-name", "Посылка", "Sort")
	addrFrom := " Россия, г. Москва, Cлавянский бульвар д.1"
	addrTo := "Россия, Воронежская обл., г. Воронеж, ул. Ленина д.43"
	size := cdek.Size{
		Height: 10,
		Length: 10,
		Weight: 4000,
		Width:  10,
	}

	array, err := post.Calculate(addrFrom, addrTo, size)
	if err != nil {
		fmt.Println(err)
	}

	var priceCending cdek.TariffCodes

	for _, ps := range array {
		if *test {
			if strings.EqualFold(ps.TariffDescription, *tariffDescription) {
				priceCending.TariffCodes = append(priceCending.TariffCodes, ps)
			} else if strings.Contains(ps.TariffName, *tariffName) {
				priceCending.TariffCodes = append(priceCending.TariffCodes, ps)
			}
		} else {
			priceCending.TariffCodes = append(priceCending.TariffCodes, ps)
		}
	}
	marshal, err := json.Marshal(&priceCending)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(marshal))
}
