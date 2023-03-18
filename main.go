package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"testProject/cdek"
)

func main() {

	test := flag.Bool("test", false, "Sort")
	tariffDescription := flag.String("tariff-description", " ", "Sort")
	tariffName := flag.String("tariff-name", "Посылка", "Sort")

	flag.Parse()

	post := cdek.NewCDEK(User, Password, ApiURL)
	if *test {
		post = cdek.NewCDEK(User, Password, ApiURLTest)
	}
	addrFrom := "270"
	addrTo := "44"
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
